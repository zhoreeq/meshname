package main

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
)

const domainZone = "meshname."

var _, validSubnet, _ = net.ParseCIDR("::/0")
var zoneConfigPath = ""
var zoneConfig = map[string][]dns.RR{}
var dnsClient = new(dns.Client)

func loadConfig() {
	if zoneConfigPath == "" {
		return
	}

	reader, err := os.Open(zoneConfigPath)
	if err != nil {
		fmt.Println("Can't open config:", err)
		return
	}

	type Zone struct {
		Domain  string
		Records []string
	}

	dec := json.NewDecoder(reader)
	for {
		var m Zone
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Syntax error in config:", err)
			return
		}
		for _, v := range m.Records {
			rr, err := dns.NewRR(v)
			if err != nil {
				fmt.Println("Invalid DNS record:", v)
				continue
			}
			zoneConfig[m.Domain] = append(zoneConfig[m.Domain], rr)
		}
	}
	fmt.Println("Config loaded:", zoneConfigPath)
}

func lookup(domain string) (net.IP, error) {
	name := strings.ToUpper(domain) + "======"
	data, err := base32.StdEncoding.DecodeString(name)
	if err != nil {
		return net.IP{}, err
	}
	if len(data) != 16 {
		return net.IP{}, errors.New("Invalid subdomain")
	}
	ipAddr := net.IP(data)
	if ipAddr == nil {
		return net.IP{}, errors.New("Invalid IP address")
	}
	if !validSubnet.Contains(ipAddr) {
		return net.IP{}, errors.New("Address from invalid subnet")
	}
	return ipAddr, nil
}

func genConf(target string) (string, error) {
	ip := net.ParseIP(target)
	if ip == nil {
		return "", errors.New("Invalid IP address")
	}
	zone := strings.ToLower(base32.StdEncoding.EncodeToString(ip)[0:26])
	selfRecord := fmt.Sprintf("\t\t\"%s.%s AAAA %s\"\n", zone, domainZone, target)
	confString := fmt.Sprintf("{\n\t\"Domain\":\"%s\",\n\t\"Records\":[\n%s\t]\n}", zone, selfRecord)

	return confString, nil
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	var remoteLookups = map[string][]dns.Question{}
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		labels := dns.SplitDomainName(q.Name)
		if len(labels) < 2 {
			continue
		}
		subDomain := labels[len(labels)-2]

		resolvedAddr, err := lookup(subDomain)
		if err != nil {
			continue
		}
		if records, ok := zoneConfig[subDomain]; ok {
			for _, rec := range records {
				if h := rec.Header(); h.Name == q.Name && h.Rrtype == q.Qtype && h.Class == q.Qclass {
					m.Answer = append(m.Answer, rec)
				}
			}
		} else if ra := w.RemoteAddr().String(); strings.HasPrefix(ra, "[::1]:") || strings.HasPrefix(ra, "127.0.0.1:") {
			 // do remote lookups only for local clients
			remoteLookups[resolvedAddr.String()] = append(remoteLookups[resolvedAddr.String()], q)
		}
	}

	for remoteServer, questions := range remoteLookups {
		rm := new(dns.Msg)
		rm.Question = questions
		resp, _, err := dnsClient.Exchange(rm, "["+remoteServer+"]:53") // no retries
		if err != nil {
			continue
		}
		m.Answer = append(m.Answer, resp.Answer...)
	}
	w.WriteMsg(m)
}

func main() {
	helpMessage := "Usage:\nmeshnamed genconf [IP] > /etc/meshnamed.conf\nmeshnamed daemon /etc/meshnamed.conf"
	if len(os.Args) < 2 {
		fmt.Println(helpMessage)
		return
	}

	action := os.Args[1]
	if action == "genconf" && len(os.Args) == 3 {
		confString, err := genConf(os.Args[2])
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(confString)
		}
	} else if action == "daemon" {
		if len(os.Args) == 3 {
			zoneConfigPath = os.Args[2]
			loadConfig()
		}

		addr := "[::1]:53535"
		if os.Getenv("LISTEN_ADDR") != "" {
			addr = os.Getenv("LISTEN_ADDR")
		}

		if os.Getenv("MESH_SUBNET") != "" {
			_, meshSubnet, err := net.ParseCIDR(os.Getenv("MESH_SUBNET"))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			validSubnet = meshSubnet
		}

		dnsClient.Timeout = 5000000000 // increased 5 seconds timeout

		dnsServer := &dns.Server{Addr: addr, Net: "udp"}
		fmt.Println("Started meshnamed on:", addr)
		dns.HandleFunc(domainZone, handleRequest)
		dnsServer.ListenAndServe()
	} else {
		fmt.Println(helpMessage)
	}
}
