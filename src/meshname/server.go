package meshname

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/gologme/log"
	"github.com/miekg/dns"
)

var DomainZones = []string{"meshname.", "ygg.", "cjd."}

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
	confString := fmt.Sprintf("{\n\t\"Domain\":\"%s\",\n\t\"Records\":[\n%s\t]\n}", subDomain, selfRecord)

	return confString, nil
}

type MeshnameServer struct {
	log                        *log.Logger
	listenAddr, zoneConfigPath string
	zoneConfig                 map[string][]dns.RR
	dnsClient                  *dns.Client
	dnsServer                  *dns.Server
	networks                   map[string]*net.IPNet
}

func (s *MeshnameServer) Init(log *log.Logger, listenAddr string, zoneConfigPath string, networks map[string]*net.IPNet) {
	s.log = log
	s.listenAddr = listenAddr
	s.networks = networks
	s.zoneConfigPath = zoneConfigPath
	s.zoneConfig = make(map[string][]dns.RR)
	if s.dnsClient == nil {
		s.dnsClient = new(dns.Client)
		s.dnsClient.Timeout = 5000000000 // increased 5 seconds timeout
	}
	s.LoadConfig()
}

func (s *MeshnameServer) LoadConfig() {
	if s.zoneConfigPath == "" {
		return
	}
	for k := range s.zoneConfig {
		delete(s.zoneConfig, k)
	}

	reader, err := os.Open(s.zoneConfigPath)
	if err != nil {
		s.log.Errorln("Can't open config:", err)
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
			s.log.Errorln("Syntax error in config:", err)
			return
		}
		for _, v := range m.Records {
			rr, err := dns.NewRR(v)
			if err != nil {
				s.log.Errorln("Invalid DNS record:", v)
				continue
			}
			s.zoneConfig[m.Domain] = append(s.zoneConfig[m.Domain], rr)
		}
	}
	s.log.Infoln("Meshname config loaded:", s.zoneConfigPath)
}

func (s *MeshnameServer) Stop() error {
	if s.dnsServer != nil {
		s.dnsServer.Shutdown()
	}
	return nil
}

func (s *MeshnameServer) Start() error {
	s.dnsServer = &dns.Server{Addr: s.listenAddr, Net: "udp"}
	for domain := range s.networks {
		dns.HandleFunc(domain, s.handleRequest)
		s.log.Debugln("Handling:", domain)
	}
	go s.dnsServer.ListenAndServe()
	s.log.Infoln("Started meshnamed on:", s.listenAddr)
	return nil
}

func (s *MeshnameServer) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	var remoteLookups = make(map[string][]dns.Question)
	m := new(dns.Msg)
	m.SetReply(r)

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

func (s *MeshnameServer) UpdateConfig() error {
	s.Stop()
	s.LoadConfig()
	s.Start()
	return nil
}
