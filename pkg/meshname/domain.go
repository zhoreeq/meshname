package meshname

import (
	"encoding/base32"
	"errors"
	"net"
	"strings"
)

func DomainFromIP(target *net.IP) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(*target)[0:26])
}

func IPFromDomain(domain *string) (net.IP, error) {
	name := strings.ToUpper(*domain) + "======"
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
	return ipAddr, nil
}
