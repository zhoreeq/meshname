package main

import (
	"fmt"
	"net"
	"strings"
	"os"

	"github.com/zhoreeq/meshname/src/meshname"
)

func main() {
	domainZone := strings.TrimSuffix(meshname.DomainZone, ".")

	usage := "Usage:\n\nmeshname lookup DOMAIN\nmeshname reverse_lookup IP"
	if len(os.Args) != 3 {
		fmt.Println(usage)
		return
	}

	action := os.Args[1]
	target := os.Args[2]

	switch action {
	case "lookup":
		labels := strings.Split(target, ".")
		if len(labels) < 2 || !strings.HasSuffix(target, domainZone) {
			fmt.Println("Invalid domain")
			return
		}
		subDomain := labels[len(labels) - 2]
		if len(subDomain) != 26 {
			fmt.Println("Invalid subdomain length")
			return
		}

		result, err := meshname.IPFromDomain(subDomain)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(result.String())
		return
	case "reverse_lookup":
		ip := net.ParseIP(target)
		if ip == nil {
			fmt.Println("Invalid IP address")
			return
		}
		result := meshname.DomainFromIP(ip)
		fmt.Println(result + "." + domainZone)
		return
	default:
		fmt.Println(usage)
		return
	}
}
