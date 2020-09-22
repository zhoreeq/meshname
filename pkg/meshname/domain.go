package meshname

import (
	"encoding/base32"
	"errors"
	"net"
	"strings"
)

// DomainFromIP derives a meshname subdomain for the authoritative DNS server address
func DomainFromIP(target *net.IP) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(*target)[0:26])
}

// IPFromDomain derives authoritative DNS server address from the meshname subdomain
func IPFromDomain(domain *string) (net.IP, error) {
	name := strings.ToUpper(*domain) + "======"
	data, err := base32.StdEncoding.DecodeString(name)
	if err != nil {
		return nil, err
	}
	if len(data) != 16 {
		return nil, errors.New("can't decode IP address, invalid subdomain")
	}
	ipAddr := net.IP(data)
	if ipAddr == nil {
		return nil, errors.New("can't decode IP address, invalid data")
	}
	return ipAddr, nil
}
