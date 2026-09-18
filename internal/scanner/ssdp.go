package scanner

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/network"
)

const ssdpSearchRequest = "M-SEARCH * HTTP/1.1\r\n" +
	"HOST: 239.255.255.250:1900\r\n" +
	"MAN: \"ssdp:discover\"\r\n" +
	"MX: 1\r\n" +
	"ST: ssdp:all\r\n" +
	"\r\n"

type SSDPDiscoverer struct {
	Timeout time.Duration
}

type ssdpResponse struct {
	RemoteIP string
	Headers  map[string]string
}

const (
	maxSSDPServiceHints    = 6
	maxSSDPServerHints     = 2
	maxSSDPDescriptorHints = 2
)

func (discoverer SSDPDiscoverer) Source() DiscoverySource {
	return DiscoverySourceSSDP
}

func (discoverer SSDPDiscoverer) Discover(ctx context.Context, local network.LocalNetwork) ([]Device, error) {
	return ssdpNeighbors(ctx, local.CIDR, discoverer.Timeout), nil
}

func ssdpNeighbors(ctx context.Context, cidr string, timeout time.Duration) []Device {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	lookupCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	responses := searchSSDP(lookupCtx, timeout)
	seen := make(map[string]Device)

	for _, response := range responses {
		ip := net.ParseIP(response.RemoteIP).To4()
		if ip == nil || !ipNet.Contains(ip) {
			continue
		}

		ipAddress := ip.String()
		device := seen[ipAddress]
		device.IP = ipAddress
		device.IsActive = true
		if device.Type == "" || device.Type == DeviceUnknown {
			device.Type = deviceTypeFromSSDP(response.Headers)
		}
		device.Hints = compactSSDPHints(appendUnique(device.Hints, ssdpHints(response.Headers)...))
		seen[ipAddress] = device
	}

	result := make([]Device, 0, len(seen))
	for _, device := range seen {
		result = append(result, device)
	}
	return result
}

func searchSSDP(ctx context.Context, timeout time.Duration) []ssdpResponse {
	target, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return nil
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil
	}
	defer conn.Close()

	if _, err := conn.WriteToUDP([]byte(ssdpSearchRequest), target); err != nil {
		return nil
	}

	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	_ = conn.SetReadDeadline(deadline)

	responses := make([]ssdpResponse, 0)
	buffer := make([]byte, 65535)
	for {
		if ctx.Err() != nil {
			return responses
		}

		n, remote, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if isNetworkTimeout(err) {
				return responses
			}
			return responses
		}
		if remote == nil || remote.IP == nil {
			continue
		}

		response := parseSSDPResponse(string(buffer[:n]))
		if len(response.Headers) == 0 {
			continue
		}
		response.RemoteIP = remote.IP.String()
		responses = append(responses, response)
	}
}

func parseSSDPResponse(payload string) ssdpResponse {
	lines := strings.Split(strings.ReplaceAll(payload, "\r\n", "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(lines[0])), "HTTP/") {
		return ssdpResponse{}
	}

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		headers[strings.ToLower(strings.TrimSpace(name))] = strings.TrimSpace(value)
	}

	return ssdpResponse{Headers: headers}
}

func deviceTypeFromSSDP(headers map[string]string) DeviceType {
	text := strings.ToLower(strings.Join([]string{
		ssdpHeader(headers, "server"),
		ssdpHeader(headers, "st"),
		ssdpHeader(headers, "usn"),
		ssdpHeader(headers, "location"),
	}, " "))

	switch {
	case strings.Contains(text, "internetgatewaydevice") || strings.Contains(text, "wanipconnection") || strings.Contains(text, "router"):
		return DeviceRouter
	case strings.Contains(text, "mediarenderer") || strings.Contains(text, "dial") || strings.Contains(text, "google") || strings.Contains(text, "cast") || strings.Contains(text, "roku") || strings.Contains(text, "hbbtv"):
		return DeviceTV
	case strings.Contains(text, "printer"):
		return DevicePrinter
	case strings.Contains(text, "camera") || strings.Contains(text, "basic:1") || strings.Contains(text, "upnp"):
		return DeviceIoT
	default:
		return DeviceUnknown
	}
}

func ssdpHints(headers map[string]string) []string {
	hints := []string{"Dispositivo anunciado por SSDP/UPnP."}

	if serviceType := ssdpHeader(headers, "st"); serviceType != "" {
		hints = append(hints, "Servicio UPnP: "+serviceType+".")
	}
	if server := ssdpHeader(headers, "server"); server != "" {
		hints = append(hints, "Servidor UPnP: "+server+".")
	}
	if location := ssdpHeader(headers, "location"); location != "" {
		hints = append(hints, "Descriptor UPnP disponible en "+location+".")
	}

	return hints
}

func compactSSDPHints(hints []string) []string {
	serviceHints := 0
	serverHints := 0
	descriptorHints := 0
	result := make([]string, 0, len(hints))

	for _, hint := range hints {
		switch {
		case strings.HasPrefix(hint, "Servicio UPnP:"):
			serviceHints++
			if serviceHints > maxSSDPServiceHints {
				continue
			}
		case strings.HasPrefix(hint, "Servidor UPnP:"):
			serverHints++
			if serverHints > maxSSDPServerHints {
				continue
			}
		case strings.HasPrefix(hint, "Descriptor UPnP disponible en"):
			descriptorHints++
			if descriptorHints > maxSSDPDescriptorHints {
				continue
			}
		}

		result = append(result, hint)
	}

	return result
}

func ssdpHeader(headers map[string]string, name string) string {
	return strings.TrimSpace(headers[strings.ToLower(name)])
}

func isNetworkTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
