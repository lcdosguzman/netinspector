package scanner

import (
	"context"
	"net"
	"os/exec"
	"regexp"

	"github.com/lcdosguzman/netinspector/internal/network"
)

var arpLinePattern = regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\) at ([0-9a-fA-F:]+)`)

type ARPDiscoverer struct{}

func (discoverer ARPDiscoverer) Source() DiscoverySource {
	return DiscoverySourceARP
}

func (discoverer ARPDiscoverer) Discover(_ context.Context, local network.LocalNetwork) ([]Device, error) {
	return arpNeighbors(local.CIDR), nil
}

func arpNeighbors(cidr string) []Device {
	output, err := exec.Command("arp", "-an").Output()
	if err != nil {
		return nil
	}

	return parseARPNeighbors(string(output), cidr)
}

func parseARPNeighbors(output string, cidr string) []Device {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	matches := arpLinePattern.FindAllStringSubmatch(output, -1)
	devices := make([]Device, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))

	for _, match := range matches {
		ip := net.ParseIP(match[1]).To4()
		if ip == nil || !ipNet.Contains(ip) || isNetworkBoundary(ip, ipNet) {
			continue
		}

		ipAddress := ip.String()
		if _, ok := seen[ipAddress]; ok {
			continue
		}
		seen[ipAddress] = struct{}{}

		deviceType := DeviceUnknown
		if ipAddress == firstHost(ipNet) {
			deviceType = DeviceRouter
		}

		devices = append(devices, Device{
			IP:       ipAddress,
			MAC:      match[2],
			Type:     deviceType,
			IsActive: true,
		})
	}

	return devices
}

func firstHost(ipNet *net.IPNet) string {
	ip := ipNet.IP.To4()
	if ip == nil {
		return ""
	}

	first := cloneIP(ip)
	incrementIP(first)
	return first.String()
}

func isNetworkBoundary(ip net.IP, ipNet *net.IPNet) bool {
	network := ipNet.IP.To4()
	if network == nil {
		return false
	}

	ip = ip.To4()
	if ip == nil {
		return false
	}

	broadcast := cloneIP(network)
	for index := range broadcast {
		broadcast[index] |= ^ipNet.Mask[index]
	}

	return ip.Equal(network) || ip.Equal(broadcast)
}
