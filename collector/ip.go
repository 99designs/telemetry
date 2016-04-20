package collector

import (
	"net"
)

func mustCIDR(cidr string) net.IPNet {
	_, net, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}

	return *net
}

var privateNetworks = []net.IPNet{
	mustCIDR("10.0.0.0/8"),
	mustCIDR("100.64.0.0/10"),
	mustCIDR("100.64.0.0/10"),
	mustCIDR("127.0.0.0/8"),
	mustCIDR("169.254.0.0/16"),
	mustCIDR("172.16.0.0/12"),
	mustCIDR("192.0.0.0/24"),
	mustCIDR("192.0.2.0/24"),
	mustCIDR("192.168.0.0/16"),
	mustCIDR("192.18.0.0/15"),
	mustCIDR("fc00::/7"),
	mustCIDR("fe80::/10"),
}

func privateIP(ip net.IP) bool {
	for _, net := range privateNetworks {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}
