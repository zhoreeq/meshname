package main

import (
	"encoding/base32"
	"fmt"
	"net"
	"strings"
	"errors"
	"os"

	"github.com/miekg/dns"
)

const domainZone = "mesh.arpa."
const maxTtl = 4294967295

var _, validSubnet, _ = net.ParseCIDR("::/0")

var srvPortMap = map[string]uint16{
	"_xmpp-client._tcp": 5222,
	"_xmpp-server._tcp": 5269,
	"_submission._tcp": 587, // rfc6186
	"_imap._tcp": 143,
	"_imaps._tcp": 993,
	"_pop3._tcp": 110,
	"_pop3s._tcp": 995,
	"_matrix._tcp": 8448, // https://matrix.org/docs/spec/server_server/unstable#server-discovery
	"_sip._tcp": 5060, // rfc3263
	"_sip._udp": 5060,
	"_sips._tcp": 5061,
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

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, v := range r.Question {
		if v.Qclass != dns.ClassINET {
			continue
		}
		labels := dns.SplitDomainName(v.Name)
		if len(labels) < 3 {
			continue
		}
		subDomain := labels[len(labels)-3]

		resolvedAddr, err := lookup(subDomain)
		if err != nil {
			continue
		}
		if v.Qtype == dns.TypeAAAA {
			r := new(dns.AAAA)
			r.Hdr = dns.RR_Header{Name: v.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: maxTtl}
			r.AAAA = resolvedAddr
			m.Answer = append(m.Answer, r)
		} else if v.Qtype == dns.TypeSRV {
			if len(labels) < 5 {
				continue
			}

			if srvRec := labels[0] + "." + labels[1]; srvPortMap[srvRec] != 0 {
				r := new(dns.SRV)
				r.Hdr = dns.RR_Header{Name: v.Name, Rrtype: dns.TypeSRV, Class: dns.ClassINET, Ttl: maxTtl}
				r.Priority = 0
				r.Weight = 0
				r.Port = srvPortMap[srvRec]
				r.Target = subDomain + "." + domainZone
				m.Answer = append(m.Answer, r)
			}
		}
	}

	w.WriteMsg(m)
}

func main() {
	addr := "127.0.0.1:53535"
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

	server := &dns.Server{Addr: addr, Net: "udp"}
	fmt.Println("Started meshnamed on:", addr)
	dns.HandleFunc(domainZone, handleRequest)
	server.ListenAndServe()
}
