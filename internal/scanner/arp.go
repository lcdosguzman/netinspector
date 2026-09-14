package scanner

import (
	"net"
	"os/exec"
	"regexp"
)

var arpLinePattern = regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\) at ([0-9a-fA-F:]+)`)

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
		if ip == nil || !ipNet.Contains(ip) {
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
