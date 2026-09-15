package scanner

import "fmt"

func mergeDevice(devices []Device, next Device) []Device {
	for index, existing := range devices {
		if existing.IP != next.IP {
			continue
		}

		if existing.MAC == "" {
			existing.MAC = next.MAC
		}
		if existing.Hostname == "" {
			existing.Hostname = next.Hostname
		}
		if existing.Vendor == "" {
			existing.Vendor = next.Vendor
		}
		if existing.Type == "" || existing.Type == DeviceUnknown {
			existing.Type = next.Type
		}
		existing.Hints = appendUnique(existing.Hints, next.Hints...)
		if existing.LatencyMS == 0 {
			existing.LatencyMS = next.LatencyMS
		}
		existing.IsActive = existing.IsActive || next.IsActive
		existing.Ports = mergePorts(existing.Ports, next.Ports)
		devices[index] = existing
		return devices
	}

	return append(devices, next)
}

func hasDevice(devices []Device, ip string) bool {
	for _, device := range devices {
		if device.IP == ip {
			return true
		}
	}
	return false
}

func mergePorts(existing []Port, next []Port) []Port {
	seen := make(map[string]struct{}, len(existing)+len(next))
	merged := make([]Port, 0, len(existing)+len(next))

	for _, port := range append(existing, next...) {
		key := fmt.Sprintf("%s/%d", port.Protocol, port.Number)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, port)
	}

	return merged
}
