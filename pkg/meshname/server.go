package meshname

import (
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/gologme/log"
	"github.com/miekg/dns"
)

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

// New is a constructor for MeshnameServer
func New(log *log.Logger, listenAddr string) *MeshnameServer {
	dnsClient := new(dns.Client)
	dnsClient.Timeout = 5000000000 // increased 5 seconds timeout

	return &MeshnameServer{
		log:        log,
		listenAddr: listenAddr,
		zoneConfig: make(map[string][]dns.RR),
		networks:   make(map[string]*net.IPNet),
		dnsClient:  dnsClient,
	}
}

func (s *MeshnameServer) Stop() {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()

	if s.started {
		if err := s.dnsServer.Shutdown(); err != nil {
			s.log.Debugln(err)
		}
		s.started = false
	}
}

func (s *MeshnameServer) Start() error {
	s.startedLock.Lock()
	defer s.startedLock.Unlock()

	if !s.started {
		s.dnsServer = &dns.Server{Addr: s.listenAddr, Net: "udp"}
		for tld, subnet := range s.networks {
			dns.HandleFunc(tld, s.handleRequest)
			s.log.Debugln("Handling:", tld, subnet)
		}
		go s.dnsServer.ListenAndServe()
		s.log.Debugln("MeshnameServer started")
		s.started = true
		return nil
	} else {
		return errors.New("MeshnameServer is already started")
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

	if err := w.WriteMsg(m); err != nil {
		s.log.Debugln("Error writing response:", err)
	}
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
