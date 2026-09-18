package scanner

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/lcdosguzman/lansweepgo/internal/network"
)

var mdnsServiceTypes = []string{
	"_workstation._tcp",
	"_http._tcp",
	"_https._tcp",
	"_airplay._tcp",
	"_raop._tcp",
	"_googlecast._tcp",
	"_companion-link._tcp",
	"_ipp._tcp",
	"_printer._tcp",
	"_smb._tcp",
	"_ssh._tcp",
}

type MDNSDiscoverer struct {
	Timeout time.Duration
}

func (discoverer MDNSDiscoverer) Source() DiscoverySource {
	return DiscoverySourceMDNS
}

func (discoverer MDNSDiscoverer) Discover(ctx context.Context, local network.LocalNetwork) ([]Device, error) {
	return mdnsNeighbors(ctx, local.CIDR, discoverer.Timeout), nil
}

func mdnsNeighbors(ctx context.Context, cidr string, timeout time.Duration) []Device {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	lookupCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry, len(mdnsServiceTypes)*8)
	var browsers sync.WaitGroup

	for _, serviceType := range mdnsServiceTypes {
		browsers.Add(1)
		go func(serviceType string) {
			defer browsers.Done()
			browseMDNSService(lookupCtx, serviceType, entries)
		}(serviceType)
	}

	go func() {
		browsers.Wait()
		close(entries)
	}()

	seen := make(map[string]Device)
	for entry := range entries {
		if entry == nil {
			continue
		}

		for _, ip := range entry.AddrIPv4 {
			if ip == nil || !ipNet.Contains(ip) {
				continue
			}

			ipAddress := ip.String()
			device := seen[ipAddress]
			device.IP = ipAddress
			device.IsActive = true
			if device.Hostname == "" {
				device.Hostname = normalizeLocalHostname(entry.HostName)
			}
			if device.Type == "" || device.Type == DeviceUnknown {
				device.Type = deviceTypeFromMDNSService(entry.Service)
			}
			device.Hints = appendUnique(device.Hints, mdnsHints(entry)...)
			seen[ipAddress] = device
		}
	}

	result := make([]Device, 0, len(seen))
	for _, device := range seen {
		result = append(result, device)
	}
	return result
}

func browseMDNSService(ctx context.Context, serviceType string, entries chan<- *zeroconf.ServiceEntry) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return
	}

	serviceEntries := make(chan *zeroconf.ServiceEntry, 32)
	go func() {
		_ = resolver.Browse(ctx, serviceType, "local.", serviceEntries)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-serviceEntries:
			if !ok {
				return
			}
			select {
			case entries <- entry:
			case <-ctx.Done():
				return
			}
		}
	}
}

func normalizeLocalHostname(hostname string) string {
	hostname = strings.TrimSpace(hostname)
	hostname = strings.TrimSuffix(hostname, ".")
	return hostname
}

func deviceTypeFromMDNSService(serviceType string) DeviceType {
	serviceType = strings.ToLower(serviceType)

	switch {
	case strings.Contains(serviceType, "_ipp") || strings.Contains(serviceType, "_printer"):
		return DevicePrinter
	case strings.Contains(serviceType, "_airplay") || strings.Contains(serviceType, "_raop") || strings.Contains(serviceType, "_googlecast") || strings.Contains(serviceType, "_companion-link"):
		return DeviceTV
	case strings.Contains(serviceType, "_workstation") || strings.Contains(serviceType, "_smb") || strings.Contains(serviceType, "_ssh"):
		return DeviceDesktop
	default:
		return DeviceUnknown
	}
}

func mdnsHints(entry *zeroconf.ServiceEntry) []string {
	hints := []string{"Hostname descubierto por mDNS/Bonjour."}

	serviceType := strings.ToLower(entry.Service)
	if strings.Contains(serviceType, "_airplay") || strings.Contains(serviceType, "_googlecast") || strings.Contains(serviceType, "_raop") {
		hints = append(hints, "Servicio multimedia anunciado por Bonjour; posible TV, speaker o dispositivo cast.")
	}
	if strings.Contains(serviceType, "_ipp") || strings.Contains(serviceType, "_printer") {
		hints = append(hints, "Servicio de impresión anunciado por Bonjour; posible impresora.")
	}
	if entry.Instance != "" {
		hints = append(hints, "Servicio anunciado: "+entry.Instance+" ("+entry.Service+").")
	}

	return hints
}

func appendUnique(existing []string, next ...string) []string {
	seen := make(map[string]struct{}, len(existing)+len(next))
	result := make([]string, 0, len(existing)+len(next))

	for _, value := range append(existing, next...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}
