package meshname

import  (
	"bytes"
	"net"
	"testing"

	"github.com/zhoreeq/meshname/src/meshname"
)

func TestIPFromDomain(t *testing.T) {
	test_subdomain := "aib7cwwdeob2vtnqf2cfnm7ilq"
	test_ip := net.ParseIP("203:f15a:c323:83aa:cdb0:2e84:56b3:e85c")

	ip, err := meshname.IPFromDomain(&test_subdomain)
	if err != nil {
		t.Errorf("Decoding IP from domain failed %s", err)
	} else if bytes.Compare(ip, test_ip) != 0 {
	   t.Errorf("Decoding IP error %s != %s", ip.String(), test_ip.String())
	}
}

func TestDomainFromIP(t *testing.T) {
	test_subdomain := "aib7cwwdeob2vtnqf2cfnm7ilq"
	test_ip := net.ParseIP("203:f15a:c323:83aa:cdb0:2e84:56b3:e85c")

	subdomain := meshname.DomainFromIP(&test_ip)
	if subdomain != test_subdomain {
		t.Errorf("Encoding domain error: %s != %s", subdomain, test_subdomain)
	}
}
