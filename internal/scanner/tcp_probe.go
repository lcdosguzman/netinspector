package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lcdosguzman/netinspector/internal/network"
)

type TCPDiscoverer struct {
	config Config
}

func NewTCPDiscoverer(config Config) TCPDiscoverer {
	return TCPDiscoverer{config: normalizeConfig(config)}
}

func (discoverer TCPDiscoverer) Source() DiscoverySource {
	return DiscoverySourceTCP
}

func (discoverer TCPDiscoverer) Discover(ctx context.Context, local network.LocalNetwork) ([]Device, error) {
	ips, err := hosts(local.CIDR)
	if err != nil {
		return nil, err
	}

	jobs := make(chan string)
	results := make(chan Device)
	var workers sync.WaitGroup

	for range discoverer.config.Concurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for ip := range jobs {
				device, ok := discoverer.probeHost(ctx, ip)
				if ok {
					results <- device
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
		close(results)
	}()

	devices := make([]Device, 0)
	for device := range results {
		devices = append(devices, device)
	}

	return devices, nil
}

func (discoverer TCPDiscoverer) probeHost(ctx context.Context, ip string) (Device, bool) {
	startedAt := time.Now()
	openPorts := make([]Port, 0)

	for _, port := range discoverer.config.Ports {
		if ctx.Err() != nil {
			return Device{}, false
		}

		address := fmt.Sprintf("%s:%d", ip, port)
		dialer := net.Dialer{Timeout: discoverer.config.Timeout}
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
