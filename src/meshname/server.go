package meshname

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"

	"github.com/gologme/log"
	"github.com/miekg/dns"
)

func DomainFromIP(target *net.IP) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(*target)[0:26])
}

func IPFromDomain(domain *string) (net.IP, error) {
	name := strings.ToUpper(*domain) + "======"
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
	return ipAddr, nil
}

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

type MeshnameServer struct {
	log        *log.Logger
	listenAddr string
	dnsClient  *dns.Client
	dnsServer  *dns.Server
	networks   map[string]*net.IPNet

	zoneConfigLock sync.RWMutex
	zoneConfig     map[string][]dns.RR

	startedLock sync.RWMutex
	started     bool
}

func (s *MeshnameServer) Init(log *log.Logger, listenAddr string) {
	s.log = log
	s.listenAddr = listenAddr
	s.zoneConfig = make(map[string][]dns.RR)
	s.networks = make(map[string]*net.IPNet)

	if s.dnsClient == nil {
		s.dnsClient = new(dns.Client)
		s.dnsClient.Timeout = 5000000000 // increased 5 seconds timeout
	}
}

func (s *MeshnameServer) Stop() error {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()

	if s.started == true {
		s.dnsServer.Shutdown()
		s.started = false
		return nil
	} else {
		return errors.New("MeshnameServer is not running")
	}
}

func (s *MeshnameServer) Start() error {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()

	if s.started == false {
		s.dnsServer = &dns.Server{Addr: s.listenAddr, Net: "udp"}
		for tld, subnet := range s.networks {
			dns.HandleFunc(tld, s.handleRequest)
			s.log.Debugln("Handling:", tld, subnet)
		}
		go s.dnsServer.ListenAndServe()
		s.log.Infoln("Started meshnamed on:", s.listenAddr)
		s.started = true
		return nil
	} else {
		return errors.New("MeshnameServer is already started")
	}
}

func (s *MeshnameServer) LoadConfig(confPath string) {
	if zoneConf, err := ParseConfigFile(confPath); err == nil {
		s.zoneConfigLock.Lock()
		s.zoneConfig = zoneConf
		s.zoneConfigLock.Unlock()
	} else {
		s.log.Errorln("Can't parse config file:", err)
	}
}

func (s *MeshnameServer) SetZoneConfig(zoneConfig map[string][]dns.RR) {
	s.zoneConfigLock.Lock()
	s.zoneConfig = zoneConfig
	s.zoneConfigLock.Unlock()
}

func (s *MeshnameServer) SetNetworks(networks map[string]*net.IPNet) {
	s.networks = networks
}

func (s *MeshnameServer) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	var remoteLookups = make(map[string][]dns.Question)
	m := new(dns.Msg)
	m.SetReply(r)

	s.zoneConfigLock.RLock()
	for _, q := range r.Question {
		labels := dns.SplitDomainName(q.Name)
		if len(labels) < 2 {
			s.log.Debugln("Error: invalid domain requested")
			continue
		}
		subDomain := labels[len(labels)-2]

		if records, ok := s.zoneConfig[subDomain]; ok {
			for _, rec := range records {
				if h := rec.Header(); h.Name == q.Name && h.Rrtype == q.Qtype && h.Class == q.Qclass {
					m.Answer = append(m.Answer, rec)
				}
			}
		} else if s.isRemoteLookupAllowed(w.RemoteAddr()) {
			// do remote lookups only for local clients
			resolvedAddr, err := IPFromDomain(&subDomain)
			if err != nil {
				s.log.Debugln(err)
				continue
			}
			// check subnet validity
			tld := labels[len(labels)-1]

			if subnet, ok := s.networks[tld]; ok && subnet.Contains(resolvedAddr) {
				remoteLookups[resolvedAddr.String()] = append(remoteLookups[resolvedAddr.String()], q)
			} else {
				s.log.Debugln("Error: subnet doesn't match")
			}
		}
	}
	s.zoneConfigLock.RUnlock()

	for remoteServer, questions := range remoteLookups {
		rm := new(dns.Msg)
		rm.Question = questions
		resp, _, err := s.dnsClient.Exchange(rm, "["+remoteServer+"]:53") // no retries
		if err != nil {
			s.log.Debugln(err)
			continue
		}
		m.Answer = append(m.Answer, resp.Answer...)
	}
	w.WriteMsg(m)
}

func (s *MeshnameServer) isRemoteLookupAllowed(addr net.Addr) bool {
	// TODO prefix whitelists ?
	ra := addr.String()
	return strings.HasPrefix(ra, "[::1]:") || strings.HasPrefix(ra, "127.0.0.1:")
}

func (s *MeshnameServer) IsStarted() bool {
	s.startedLock.RLock()
	started := s.started
	s.startedLock.RUnlock()
	return started
}
