package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"os/signal"
	"syscall"

	"github.com/gologme/log"

	"github.com/zhoreeq/meshname/src/meshname"
)

func main() {
	genconf := flag.String("genconf", "", "generate a new config for IP address")
	subdomain := flag.String("subdomain", "meshname.", "subdomain used to generate config")
	useconffile := flag.String("useconffile", "", "run daemon with a config file")
	listenAddr := flag.String("listenaddr", "[::1]:53535", "address to listen on")
	networksconf := flag.String("networks", "ygg=200::/7,cjd=fc00::/8,meshname=::/0", "TLD=subnet list separated by comma")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	var logger *log.Logger
	logger = log.New(os.Stdout, "", log.Flags())

	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")
	if *debug {
		logger.EnableLevel("debug")
	}

	if *genconf != "" {
		confString, err := meshname.GenConf(*genconf, *subdomain)
		if err != nil {
			logger.Errorln(err)
		} else {
			fmt.Println(confString)
		}
		return
	}

	networks := make(map[string]string)
	for _, item := range strings.Split(*networksconf, ",") {
		if tokens := strings.SplitN(item, "=", 2); len(tokens) == 2 {
			networks[tokens[0]] = tokens[1]
		}
	}

	s := new(meshname.MeshnameServer)
	s.Init(logger, *listenAddr, *useconffile, networks)
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
			s.UpdateConfig()
		}
	}
}
