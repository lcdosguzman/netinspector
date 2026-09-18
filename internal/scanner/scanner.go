package scanner

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/network"
)

type Config struct {
	Ports       []int
	Concurrency int
	Timeout     time.Duration
}

type TCPScanner struct {
	config      Config
	discoverers []Discoverer
}

func NewTCPScanner(config Config) TCPScanner {
	config = normalizeConfig(config)
	return TCPScanner{
		config:      config,
		discoverers: defaultDiscoverers(config),
	}
}

func normalizeConfig(config Config) Config {
	if config.Concurrency <= 0 {
		config.Concurrency = 64
	}
	if config.Timeout <= 0 {
		config.Timeout = 300 * time.Millisecond
	}
	if len(config.Ports) == 0 {
		config.Ports = DefaultDiscoveryPorts()
	}

	return config
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

	for _, discoverer := range scanner.activeDiscoverers() {
		devices, err := discoverer.Discover(ctx, local)
		if err != nil {
			return ScanResult{}, err
		}

		for _, device := range devices {
			alreadyKnown := hasDevice(result.Devices, device.IP)
			result.Devices = mergeDevice(result.Devices, device)
			if shouldEmitDiscoveryEvent(discoverer.Source(), alreadyKnown) {
				result.Events = append(result.Events, ScanEvent{
					Type:      "DEVICE_DISCOVERED",
					Message:   discoveryEventMessage(discoverer.Source(), device.IP),
					DeviceIP:  device.IP,
					Timestamp: time.Now().UTC(),
				})
			}
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

func (scanner TCPScanner) activeDiscoverers() []Discoverer {
	if len(scanner.discoverers) > 0 {
		return scanner.discoverers
	}
	return defaultDiscoverers(normalizeConfig(scanner.config))
}
