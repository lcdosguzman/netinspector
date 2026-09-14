package scanner

import (
	"context"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/lcdosguzman/netinspector/internal/network"
)

type Config struct {
	Ports       []int
	Concurrency int
	Timeout     time.Duration
}

type TCPScanner struct {
	config Config
}

func NewTCPScanner(config Config) TCPScanner {
	if config.Concurrency <= 0 {
		config.Concurrency = 64
	}
	if config.Timeout <= 0 {
		config.Timeout = 300 * time.Millisecond
	}
	if len(config.Ports) == 0 {
		config.Ports = DefaultDiscoveryPorts()
	}

	return TCPScanner{config: config}
}

func DefaultDiscoveryPorts() []int {
	return []int{22, 53, 80, 139, 443, 445, 631, 1900, 3389, 5900, 8000, 8080, 8443}
}

func (scanner TCPScanner) ScanCIDR(ctx context.Context, local network.LocalNetwork) (ScanResult, error) {
	startedAt := time.Now().UTC()
	ips, err := hosts(local.CIDR)
	if err != nil {
		return ScanResult{}, err
	}

	result := ScanResult{
		ID:        fmt.Sprintf("scan-%d", startedAt.UnixNano()),
		Mode:      "REAL",
		Network:   local.CIDR,
		StartedAt: startedAt,
		Events: []ScanEvent{{
			Type:      "SCAN_STARTED",
			Message:   fmt.Sprintf("Scanning %d hosts on %s", len(ips), local.CIDR),
			Timestamp: startedAt,
		}},
	}

	jobs := make(chan string)
	devices := make(chan Device)
	var workers sync.WaitGroup

	for range scanner.config.Concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for ip := range jobs {
				device, ok := scanner.probeHost(ctx, ip)
				if ok {
					devices <- device
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, ip := range ips {
			select {
			case <-ctx.Done():
				return
			case jobs <- ip:
			}
		}
	}()

	go func() {
		workers.Wait()
		close(devices)
	}()

	for device := range devices {
		result.Devices = mergeDevice(result.Devices, device)
		result.Events = append(result.Events, ScanEvent{
			Type:      "DEVICE_DISCOVERED",
			Message:   fmt.Sprintf("Discovered %s", device.IP),
			DeviceIP:  device.IP,
			Timestamp: time.Now().UTC(),
		})
	}

	for _, device := range arpNeighbors(local.CIDR) {
		alreadyKnown := hasDevice(result.Devices, device.IP)
		result.Devices = mergeDevice(result.Devices, device)
		if !alreadyKnown {
			result.Events = append(result.Events, ScanEvent{
				Type:      "DEVICE_DISCOVERED",
				Message:   fmt.Sprintf("Discovered %s from ARP cache", device.IP),
				DeviceIP:  device.IP,
				Timestamp: time.Now().UTC(),
			})
		}
	}

	for index, device := range result.Devices {
		result.Devices[index] = enrichDevice(device)
	}

	sort.Slice(result.Devices, func(i, j int) bool {
		return compareIPv4(result.Devices[i].IP, result.Devices[j].IP) < 0
	})

	result.EndedAt = time.Now().UTC()
	result.Events = append(result.Events, ScanEvent{
		Type:      "SCAN_FINISHED",
		Message:   fmt.Sprintf("Scan finished with %d active devices", len(result.Devices)),
		Timestamp: result.EndedAt,
	})

	return result, ctx.Err()
}

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

func (scanner TCPScanner) probeHost(ctx context.Context, ip string) (Device, bool) {
	startedAt := time.Now()
	openPorts := make([]Port, 0)

	for _, port := range scanner.config.Ports {
		if ctx.Err() != nil {
			return Device{}, false
		}

		address := fmt.Sprintf("%s:%d", ip, port)
		dialer := net.Dialer{Timeout: scanner.config.Timeout}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			continue
		}
		_ = conn.Close()

		openPorts = append(openPorts, Port{
			Number:      port,
			Protocol:    "tcp",
			ServiceName: serviceName(port),
		})
	}

	if len(openPorts) == 0 {
		return Device{}, false
	}

	return Device{
		IP:        ip,
		Type:      DeviceUnknown,
		LatencyMS: time.Since(startedAt).Milliseconds(),
		IsActive:  true,
		Ports:     openPorts,
	}, true
}

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

func serviceName(port int) string {
	names := map[int]string{
		22:   "ssh",
		53:   "dns",
		80:   "http",
		139:  "netbios",
		443:  "https",
		445:  "smb",
		631:  "ipp",
		1900: "ssdp",
		3389: "rdp",
		5900: "vnc",
		8000: "http-alt",
		8080: "http-proxy",
		8443: "https-alt",
	}
	if name, ok := names[port]; ok {
		return name
	}
	return "unknown"
}
