package meshname

import (
	"net"
	"os"
	"strings"
	"testing"

	"github.com/gologme/log"
	"github.com/miekg/dns"

	"github.com/zhoreeq/meshname/pkg/meshname"
)

func TestServerLocalDomain(t *testing.T) {
	bindAddr := "[::1]:54545"
	log := log.New(os.Stdout, "", log.Flags())

	ts := meshname.New(log, bindAddr)
	// ...
	yggIPNet := &net.IPNet{IP: net.ParseIP("200::"), Mask: net.CIDRMask(7, 128)}
	ts.SetNetworks(map[string]*net.IPNet{"ygg": yggIPNet, "meshname": yggIPNet})

	exampleConfig := make(map[string][]string)
	exampleConfig["aiarnf2wpqjxkp6rhivuxbondy"] = append(exampleConfig["aiarnf2wpqjxkp6rhivuxbondy"],
		"test.aiarnf2wpqjxkp6rhivuxbondy.meshname. AAAA 201:1697:567c:1375:3fd1:3a2b:4b85:cd1e")

	if zoneConfig, err := meshname.ParseZoneConfigMap(exampleConfig); err == nil {
		ts.SetZoneConfig(zoneConfig)
	} else {
		t.Fatalf("meshname: Failed to parse Meshname config: %s", err)
	}

	ts.Start()

	tc := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion("test.aiarnf2wpqjxkp6rhivuxbondy.meshname.", dns.TypeAAAA)
	resp, _, err := tc.Exchange(m, bindAddr)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("Invalid response: %s", resp.String())
	}
	if !strings.Contains(resp.Answer[0].String(), "201:1697:567c:1375:3fd1:3a2b:4b85:cd1e") {
		t.Fatalf("Invalid IP in response: %s", resp.Answer[0].String())
	}

	ts.Stop()
}
