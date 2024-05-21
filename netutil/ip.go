package netutil

import (
	"net"
)

func IsPrivateIp(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // 链路本地地址
		"127.0.0.0/8",    // 本地环回地址
		"127.0.0.1/32",
		"::1/128",   // IPv6本地环回地址
		"fe80::/10", // IPv6链路本地地址
	}
	for _, r := range privateRanges {
		_, network, _ := net.ParseCIDR(r)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
