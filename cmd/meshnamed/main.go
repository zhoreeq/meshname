package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gologme/log"

	"github.com/zhoreeq/meshname/pkg/meshname"
)

func parseNetworks(networksconf string) (map[string]*net.IPNet, error) {
	networks := make(map[string]*net.IPNet)
	for _, item := range strings.Split(networksconf, ",") {
		if tokens := strings.SplitN(item, "=", 2); len(tokens) == 2 {
			if _, validSubnet, err := net.ParseCIDR(tokens[1]); err == nil {
				networks[tokens[0]] = validSubnet
			} else {
				return nil, err
			}
		}
	}
	return networks, nil
}

var (
	listenAddr, networksconf string
	getName, getIP           string
	debug, noMeshIP          bool
)

func init() {
	flag.StringVar(&listenAddr, "listenaddr", "[::1]:53535", "address to listen on")
	flag.StringVar(&networksconf, "networks", "ygg=200::/7,cjd=fc00::/8,meshname=::/0,popura=::/0", "TLD=subnet list separated by comma")
	flag.BoolVar(&noMeshIP, "nomeship", false, "disable .meship resolver")
	flag.StringVar(&getName, "getname", "", "convert IPv6 address to a name")
	flag.StringVar(&getIP, "getip", "", "convert a name to IPv6 address")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
}

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Flags())

	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")
	if debug {
		logger.EnableLevel("debug")
	}

	if getName != "" {
		ip := net.ParseIP(getName)
		if ip == nil {
			logger.Fatal("Invalid IP address")
		}
		subDomain := meshname.DomainFromIP(&ip)
		fmt.Println(subDomain)
		return
	} else if getIP != "" {
		ip, err := meshname.IPFromDomain(&getIP)
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println(ip)
		return
	}

	networks, err := parseNetworks(networksconf)
	if err != nil {
		logger.Fatalln(err)
	}

	s := meshname.New(logger, listenAddr, networks, !noMeshIP)

	if err := s.Start(); err != nil {
		logger.Fatal(err)
	}
	logger.Infoln("Listening on:", listenAddr)

	c := make(chan os.Signal, 1)
	r := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(r, os.Interrupt, syscall.SIGHUP)
	defer s.Stop()
	for {
		select {
		case <-c:
			return
		}
	}
}
