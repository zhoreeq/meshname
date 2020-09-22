package meshname

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/miekg/dns"
)


func GenConf(target, zone string) (string, error) {
	ip := net.ParseIP(target)
	if ip == nil {
		return "", errors.New("Invalid IP address")
	}
	subDomain := DomainFromIP(&ip)
	selfRecord := fmt.Sprintf("\t\t\"%s.%s AAAA %s\"\n", subDomain, zone, target)
	confString := fmt.Sprintf("{\n\t\"%s\":[\n%s\t]\n}", subDomain, selfRecord)

	return confString, nil
}

// Load zoneConfig from a JSON file
func ParseConfigFile(configPath string) (map[string][]dns.RR, error) {
	conf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var dat map[string][]string
	if err := json.Unmarshal(conf, &dat); err == nil {
		return ParseZoneConfigMap(dat)
	} else {
		return nil, err
	}
}

func ParseZoneConfigMap(zoneConfigMap map[string][]string) (map[string][]dns.RR, error) {
	var zoneConfig = make(map[string][]dns.RR)
	for subDomain, records := range zoneConfigMap {
		for _, r := range records {
			if rr, err := dns.NewRR(r); err == nil {
				zoneConfig[subDomain] = append(zoneConfig[subDomain], rr)
			} else {
				return nil, err
			}
		}
	}
	return zoneConfig, nil
}

