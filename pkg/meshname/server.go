package meshname

import (
	"errors"
	"net"
	"sync"

	"github.com/gologme/log"
	"github.com/miekg/dns"
)

type MeshnameServer struct {
	log          *log.Logger
	listenAddr   string
	dnsClient    *dns.Client
	dnsServer    *dns.Server
	networks     map[string]*net.IPNet
	enableMeshIP bool

	startedLock sync.RWMutex
	started     bool
}

// New is a constructor for MeshnameServer
func New(log *log.Logger, listenAddr string, networks map[string]*net.IPNet, enableMeshIP bool) *MeshnameServer {
	dnsClient := new(dns.Client)
	dnsClient.Timeout = 5000000000 // increased 5 seconds timeout

	return &MeshnameServer{
		log:          log,
		listenAddr:   listenAddr,
		networks:     networks,
		dnsClient:    dnsClient,
		enableMeshIP: enableMeshIP,
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
			Addr:              s.listenAddr,
			Net:               "udp",
			NotifyStartedFunc: func() { close(waitStarted) },
		}
		for tld, subnet := range s.networks {
			dns.HandleFunc(tld, s.handleMeshnameRequest)
			s.log.Debugln("Handling:", tld, subnet)
		}
		if s.enableMeshIP {
			dns.HandleFunc("meship", s.handleMeshIPRequest)
			s.log.Debugln("Handling: meship ::/0")
		}

		go func() {
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

func (s *MeshnameServer) handleMeshnameRequest(w dns.ResponseWriter, r *dns.Msg) {
	var remoteLookups = make(map[string][]dns.Question)
	m := new(dns.Msg)
	m.SetReply(r)
	s.log.Debugln(r.String())

	for _, q := range r.Question {
		labels := dns.SplitDomainName(q.Name)
		if len(labels) < 2 {
			s.log.Debugln("Error: invalid domain requested")
			continue
		}
		subDomain := labels[len(labels)-2]

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

	for remoteServer, questions := range remoteLookups {
		rm := new(dns.Msg)
		rm.RecursionDesired = true
		rm.Question = questions
		resp, _, err := s.dnsClient.Exchange(rm, "["+remoteServer+"]:53") // no retries
		if err != nil {
			s.log.Debugln(err)
			continue
		}
		s.log.Debugln(resp.String())
		m.Answer = append(m.Answer, resp.Answer...)
		m.Ns = append(m.Ns, resp.Ns...)
		m.Extra = append(m.Extra, resp.Extra...)
	}

	if err := w.WriteMsg(m); err != nil {
		s.log.Debugln("Error writing response:", err)
	}
}

func (s *MeshnameServer) handleMeshIPRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		labels := dns.SplitDomainName(q.Name)
		// resolve only 2nd level domains and AAAA type
		if len(labels) != 2 || q.Qtype != dns.TypeAAAA || q.Qclass != dns.ClassINET {
			s.log.Debugln("Error: invalid resource requested")
			continue
		}

		if resolvedAddr, err := IPFromDomain(&labels[0]); err == nil {
			answer := new(dns.AAAA)
			answer.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 3600}
			answer.AAAA = resolvedAddr

			m.Answer = append(m.Answer, answer)
		} else {
			s.log.Debugln(err)
		}
	}

	if err := w.WriteMsg(m); err != nil {
		s.log.Debugln("Error writing response:", err)
	}
}

func (s *MeshnameServer) IsStarted() bool {
	s.startedLock.RLock()
	started := s.started
	s.startedLock.RUnlock()
	return started
}
