package collector

import (
	"testing"
	"net"
)

func TestPrivateIp(t *testing.T) {
	checks := map[string]bool {
		"192.168.1.1": true,
		"8.8.8.8": false,
		"172.16.100.0": true,
		"172.32.1.1": false,
		"10.0.0.1": true,
		"10.255.255.254": true,
		"dead:beef:cafe:f00d:dead:beef:cafe:f00d": false,
		"fc00:beef:cafe:f00d:dead:beef:cafe:f00d": true,
	}

	for ip, isPrivate := range checks {
		if privateIP(net.ParseIP(ip)) != isPrivate {
			str := "not private"
			if !isPrivate {
				str = "private"
			}
			t.Errorf("%s is icorrectly marked as %s", ip, str)
		}
	}
}
