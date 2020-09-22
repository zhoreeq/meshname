package meshname

import  (
	"bytes"
	"net"
	"testing"
	"fmt"

	"github.com/zhoreeq/meshname/pkg/meshname"
)

func TestIPFromDomain(t *testing.T) {
	test_subdomain := "aib7cwwdeob2vtnqf2cfnm7ilq"
	test_ip := net.ParseIP("203:f15a:c323:83aa:cdb0:2e84:56b3:e85c")

	if ip, err := meshname.IPFromDomain(&test_subdomain); err != nil {
		t.Fatal(err)
	} else if bytes.Compare(ip, test_ip) != 0 {
	   t.Fatalf("Decoding IP error %s != %s", ip.String(), test_ip.String())
	}
}

func TestDomainFromIP(t *testing.T) {
	test_subdomain := "aib7cwwdeob2vtnqf2cfnm7ilq"
	test_ip := net.ParseIP("203:f15a:c323:83aa:cdb0:2e84:56b3:e85c")

	subdomain := meshname.DomainFromIP(&test_ip)
	if subdomain != test_subdomain {
		t.Fatalf("Encoding domain error: %s != %s", subdomain, test_subdomain)
	}
}

func ExampleIPFromDomain() {
	test_subdomain := "aib7cwwdeob2vtnqf2cfnm7ilq"

	if ip, err := meshname.IPFromDomain(&test_subdomain); err == nil {
		fmt.Println(ip)
	}
	// Output: 203:f15a:c323:83aa:cdb0:2e84:56b3:e85c
}

func ExampleDomainFromIP() {
	test_ip := net.ParseIP("203:f15a:c323:83aa:cdb0:2e84:56b3:e85c")

	fmt.Println(meshname.DomainFromIP(&test_ip))
	// Output: aib7cwwdeob2vtnqf2cfnm7ilq
}
