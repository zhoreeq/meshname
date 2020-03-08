package main

import (
	"net"
	"os"
	"os/signal"
	"fmt"
	"flag"
	"syscall"

	"github.com/gologme/log"

	"github.com/zhoreeq/meshname/src/meshname"
)


func main() {
	genconf := flag.String("genconf", "", "generate a new config for IP address")
	useconffile := flag.String("useconffile", "", "run daemon with a config file")
	listenAddr := flag.String("listenaddr", "[::1]:53535", "address to listen on")
	meshSubnetStr := flag.String("meshsubnet", "::/0", "valid IPv6 address space")
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

	switch {
	case *genconf != "":
		confString, err := meshname.GenConf(*genconf)
		if err != nil {
			logger.Errorln(err)
		} else {
			fmt.Println(confString)
		}
	case *useconffile != "":
		s := new(meshname.MeshnameServer)

		_, validSubnet, err := net.ParseCIDR(*meshSubnetStr)
		if err != nil {
			logger.Errorln(err)
			os.Exit(1)
		}

		s.Init(logger, *listenAddr, *useconffile, validSubnet)
		s.Start()

		c := make(chan os.Signal, 1)
		r := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		signal.Notify(r, os.Interrupt, syscall.SIGHUP)
		defer s.Stop()
		for {
			select {
			case _ = <-c:
				goto exit
			case _ = <-r:
				s.UpdateConfig()
			}
		}
	default:
		flag.PrintDefaults()
	}
exit:
}
