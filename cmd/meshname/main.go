package main

import (
	"encoding/base32"
	"fmt"
	"net"
	"strings"
	"os"
	"errors"
)

var domainZone = ".meshname"

func reverse_lookup(target string) (string, error) {
	ip := net.ParseIP(target)
	if ip == nil {
		return "", errors.New("Invalid IP address")
	}
	str := base32.StdEncoding.EncodeToString(ip)[0:26]
	return strings.ToLower(str) + domainZone, nil
}

func lookup(target string)  (string, error) {
	labels := strings.Split(target, ".")
	if len(labels) < 3 || strings.HasSuffix(domainZone, target) {
		return "", errors.New("Invalid domain")
	}
	subDomain := labels[len(labels) - 3]
	if len(subDomain) != 26 {
		return "", errors.New("Invalid subdomain length")
	}
	name := strings.ToUpper(subDomain) + "======"
	data, err := base32.StdEncoding.DecodeString(name)
	if err != nil {
		return "", err
	}
	s := net.IP(data)
	if s == nil {
		return "", errors.New("Invalid IP address")
	}
	return s.String(), nil
}

func main() {
	usage := "Usage:\n\nmeshname lookup DOMAIN\nmeshname reverse_lookup IP"
	if len(os.Args) != 3 {
		fmt.Println(usage)
		return
	}

	action := os.Args[1]
	target := os.Args[2]

	switch action {
	case "lookup":
		result, err := lookup(target)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(result)
		return
	case "reverse_lookup":
		result, err := reverse_lookup(target)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(result)
		return
	default:
		fmt.Println(usage)
		return
	}
}
