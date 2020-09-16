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

	"github.com/zhoreeq/meshname/src/meshname"
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
	genconf, subdomain, useconffile, listenAddr, networksconf string
	debug bool
)

func init() {
	flag.StringVar(&genconf, "genconf", "", "generate a new config for IP address")
	flag.StringVar(&subdomain, "subdomain", "meshname.", "subdomain used to generate config")
	flag.StringVar(&useconffile, "useconffile", "", "run daemon with a config file")
	flag.StringVar(&listenAddr, "listenaddr", "[::1]:53535", "address to listen on")
	flag.StringVar(&networksconf, "networks", "ygg=200::/7,cjd=fc00::/8,meshname=::/0", "TLD=subnet list separated by comma")
	flag.BoolVar(&debug,"debug", false, "enable debug logging")
}

func main() {
	flag.Parse()

	var logger *log.Logger
	logger = log.New(os.Stdout, "", log.Flags())

	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")
	if debug {
		logger.EnableLevel("debug")
	}

	if genconf != "" {
		if conf, err := meshname.GenConf(genconf, subdomain); err == nil {
			fmt.Println(conf)
		} else {
			logger.Errorln(err)
		}
		return
	}


	s := new(meshname.MeshnameServer)
	s.Init(logger, listenAddr)

	if networks, err := parseNetworks(networksconf); err == nil {
		s.SetNetworks(networks)
	} else {
		logger.Errorln(err)
	}

	if useconffile != "" {
		s.LoadConfig(useconffile)
	}

	s.Start()

	c := make(chan os.Signal, 1)
	r := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(r, os.Interrupt, syscall.SIGHUP)
	defer s.Stop()
	for {
		select {
		case _ = <-c:
			return
		case _ = <-r:
			if useconffile != "" {
				s.LoadConfig(useconffile)
			}
		}
	}
}
