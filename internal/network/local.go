package network

import (
	"errors"
	"fmt"
	"net"
)

type LocalNetwork struct {
	InterfaceName string `json:"interfaceName"`
	IP            string `json:"ip"`
	CIDR          string `json:"cidr"`
}

func DetectLocalNetwork() (LocalNetwork, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return LocalNetwork{}, fmt.Errorf("list network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if !isUsableInterface(iface) {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, address := range addresses {
			ip, network, ok := ipv4Network(address)
			if !ok {
				continue
			}

			return LocalNetwork{
				InterfaceName: iface.Name,
				IP:            ip.String(),
				CIDR:          network.String(),
			}, nil
		}
	}

	return LocalNetwork{}, errors.New("no active non-loopback IPv4 network interface found")
}

func isUsableInterface(iface net.Interface) bool {
	if iface.Flags&net.FlagUp == 0 {
		return false
	}
	if iface.Flags&net.FlagLoopback != 0 {
		return false
	}
	return true
}

func ipv4Network(address net.Addr) (net.IP, *net.IPNet, bool) {
	ipNet, ok := address.(*net.IPNet)
	if !ok {
		return nil, nil, false
	}

	ip := ipNet.IP.To4()
	if ip == nil {
		return nil, nil, false
	}

	networkIP := make(net.IP, len(ip))
	copy(networkIP, ip)
	networkIP = networkIP.Mask(ipNet.Mask)

	return ip, &net.IPNet{IP: networkIP, Mask: ipNet.Mask}, true
}
