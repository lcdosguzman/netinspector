package scanner

import (
	"fmt"
	"net"
)

func hosts(cidr string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("parse cidr %q: %w", cidr, err)
	}

	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("cidr %q is not an IPv4 network", cidr)
	}

	var addresses []string
	for current := cloneIP(ip.Mask(ipNet.Mask)); ipNet.Contains(current); incrementIP(current) {
		addresses = append(addresses, current.String())
	}

	if len(addresses) <= 2 {
		return addresses, nil
	}

	return addresses[1 : len(addresses)-1], nil
}

func cloneIP(ip net.IP) net.IP {
	clone := make(net.IP, len(ip))
	copy(clone, ip)
	return clone
}

func incrementIP(ip net.IP) {
	for index := len(ip) - 1; index >= 0; index-- {
		ip[index]++
		if ip[index] != 0 {
			break
		}
	}
}

func compareIPv4(left string, right string) int {
	leftIP := net.ParseIP(left).To4()
	rightIP := net.ParseIP(right).To4()
	if leftIP == nil || rightIP == nil {
		if left < right {
			return -1
		}
		if left > right {
			return 1
		}
		return 0
	}

	for index := range leftIP {
		if leftIP[index] < rightIP[index] {
			return -1
		}
		if leftIP[index] > rightIP[index] {
			return 1
		}
	}
	return 0
}
