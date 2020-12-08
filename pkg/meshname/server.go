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
	allowRemote bool

	dnsRecordsLock sync.RWMutex
	dnsRecords     map[string][]dns.RR

	startedLock sync.RWMutex
	started     bool
}

// New is a constructor for MeshnameServer
func New(log *log.Logger, listenAddr string, networks map[string]*net.IPNet, allowRemote bool) *MeshnameServer {
	dnsClient := new(dns.Client)
	dnsClient.Timeout = 5000000000 // increased 5 seconds timeout

	return &MeshnameServer{
		log:        log,
		listenAddr: listenAddr,
		dnsRecords: make(map[string][]dns.RR),
		networks:   networks,
		dnsClient:  dnsClient,
		allowRemote: allowRemote,
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
		waitStarted := make(chan struct{})
		s.dnsServer = &dns.Server{
			Addr: s.listenAddr,
			Net: "udp",
			NotifyStartedFunc: func(){ close(waitStarted) },
		}
		for tld, subnet := range s.networks {
			dns.HandleFunc(tld, s.handleRequest)
			s.log.Debugln("Handling:", tld, subnet)
		}
		go func(){
			if err := s.dnsServer.ListenAndServe(); err != nil {
				s.log.Fatalln("MeshnameServer failed to start:", err)
			}
		}()
		<-waitStarted

		s.log.Debugln("MeshnameServer started")
		s.started = true
		return nil
	} else {
		return errors.New("MeshnameServer is already started")
	}
}

func (s *MeshnameServer) ConfigureDNSRecords(dnsRecords map[string][]dns.RR) {
	s.dnsRecordsLock.Lock()
	s.dnsRecords = dnsRecords
	s.dnsRecordsLock.Unlock()
}

func (s *MeshnameServer) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	var remoteLookups = make(map[string][]dns.Question)
	m := new(dns.Msg)
	m.SetReply(r)

	s.dnsRecordsLock.RLock()
	for _, q := range r.Question {
		labels := dns.SplitDomainName(q.Name)
		if len(labels) < 2 {
			s.log.Debugln("Error: invalid domain requested")
			continue
		}
		subDomain := labels[len(labels)-2]

		if records, ok := s.dnsRecords[subDomain]; ok {
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
	s.dnsRecordsLock.RUnlock()

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
	if s.allowRemote {
		return true
	}
	ra := addr.String()
	return strings.HasPrefix(ra, "[::1]:") || strings.HasPrefix(ra, "127.0.0.1:")
}

func (s *MeshnameServer) IsStarted() bool {
	s.startedLock.RLock()
	started := s.started
	s.startedLock.RUnlock()
	return started
}
