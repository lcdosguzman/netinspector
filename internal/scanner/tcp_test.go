package scanner

import (
	"context"
	"reflect"
	"testing"

	"github.com/lcdosguzman/netinspector/internal/network"
)

type fakeDiscoverer struct {
	source  DiscoverySource
	devices []Device
}

func (discoverer fakeDiscoverer) Source() DiscoverySource {
	return discoverer.source
}

func (discoverer fakeDiscoverer) Discover(_ context.Context, _ network.LocalNetwork) ([]Device, error) {
	return discoverer.devices, nil
}

func TestHostsSkipsNetworkAndBroadcast(t *testing.T) {
	got, err := hosts("192.168.1.0/30")
	if err != nil {
		t.Fatalf("hosts returned error: %v", err)
	}

	want := []string{"192.168.1.1", "192.168.1.2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hosts() = %v, want %v", got, want)
	}
}

func TestScanCIDRUsesConfiguredDiscoverers(t *testing.T) {
	scanner := TCPScanner{
		config: normalizeConfig(Config{}),
		discoverers: []Discoverer{
			fakeDiscoverer{
				source: DiscoverySourceTCP,
				devices: []Device{{
					IP:       "192.168.1.2",
					Type:     DeviceUnknown,
					IsActive: true,
				}},
			},
			fakeDiscoverer{
				source: DiscoverySourceARP,
				devices: []Device{{
					IP:       "192.168.1.3",
					MAC:      "84:15:d3:1b:8e:dd",
					Type:     DeviceUnknown,
					IsActive: true,
				}},
			},
		},
	}

	result, err := scanner.ScanCIDR(context.Background(), network.LocalNetwork{CIDR: "192.168.1.0/30"})
	if err != nil {
		t.Fatalf("ScanCIDR returned error: %v", err)
	}

	if len(result.Devices) != 2 {
		t.Fatalf("ScanCIDR returned %d devices, want 2", len(result.Devices))
	}
	if result.Events[1].Message != "Discovered 192.168.1.2" {
		t.Fatalf("TCP discovery event = %q, want default discovery message", result.Events[1].Message)
	}
	if result.Events[2].Message != "Discovered 192.168.1.3 from ARP cache" {
		t.Fatalf("ARP discovery event = %q, want ARP discovery message", result.Events[2].Message)
	}
}

func TestServiceName(t *testing.T) {
	tests := map[int]string{
		22:    "ssh",
		443:   "https",
		65535: "unknown",
	}

	for port, want := range tests {
		if got := serviceName(port); got != want {
			t.Fatalf("serviceName(%d) = %q, want %q", port, got, want)
		}
	}
}

func TestParseARPNeighbors(t *testing.T) {
	output := `? (192.168.1.1) at 84:15:d3:1b:8e:dd on en0 ifscope [ethernet]
? (192.168.1.7) at 78:66:9d:b:bc:16 on en0 ifscope [ethernet]
? (192.168.1.8) at (incomplete) on en0 ifscope [ethernet]
? (192.168.1.255) at ff:ff:ff:ff:ff:ff on en0 ifscope [ethernet]
? (224.0.0.251) at 1:0:5e:0:0:fb on en0 ifscope permanent [ethernet]`

	got := parseARPNeighbors(output, "192.168.1.0/24")
	if len(got) != 2 {
		t.Fatalf("parseARPNeighbors() returned %d devices, want 2", len(got))
	}

	if got[0].IP != "192.168.1.1" || got[0].Type != DeviceRouter {
		t.Fatalf("first ARP neighbor = %+v, want router 192.168.1.1", got[0])
	}
	if got[1].IP != "192.168.1.7" || got[1].MAC != "78:66:9d:b:bc:16" {
		t.Fatalf("second ARP neighbor = %+v, want 192.168.1.7 with MAC", got[1])
	}
}

func TestVendorNameDetectsLocallyAdministeredMAC(t *testing.T) {
	got := vendorName("9e:68:c4:89:82:06", "")
	if got != "Private/randomized MAC" {
		t.Fatalf("vendorName() = %q, want private/randomized marker", got)
	}
}

func TestEnrichDeviceUsesOUIAndTypeHints(t *testing.T) {
	device := enrichDevice(Device{
		IP:       "192.168.1.7",
		MAC:      "78:66:9d:0b:bc:16",
		Type:     DeviceUnknown,
		IsActive: true,
		Ports: []Port{
			{Number: 8443, Protocol: "tcp", ServiceName: "https-alt"},
		},
	})

	if device.Vendor != "Hui Zhou Gaoshengda Technology Co., Ltd." {
		t.Fatalf("Vendor = %q, want Hui Zhou Gaoshengda", device.Vendor)
	}
	if device.Type != DeviceTV {
		t.Fatalf("Type = %q, want %q", device.Type, DeviceTV)
	}
	if len(device.Hints) == 0 {
		t.Fatal("expected identification hints")
	}
}

func TestEnrichDevicePrefersComputerHostnameOverAirPlayService(t *testing.T) {
	device := enrichDevice(Device{
		IP:       "192.168.1.11",
		MAC:      "3c:06:30:4c:e0:d1",
		Hostname: "Simons-MacBook-Pro.local",
		Type:     DeviceTV,
		IsActive: true,
		Hints: []string{
			"Servicio multimedia anunciado por Bonjour; posible TV, speaker o dispositivo cast.",
		},
	})

	if device.Type != DeviceDesktop {
		t.Fatalf("Type = %q, want %q", device.Type, DeviceDesktop)
	}
}

func TestDeviceTypeFromMDNSService(t *testing.T) {
	tests := map[string]DeviceType{
		"_ipp._tcp":         DevicePrinter,
		"_airplay._tcp":     DeviceTV,
		"_googlecast._tcp":  DeviceTV,
		"_workstation._tcp": DeviceDesktop,
	}

	for serviceType, want := range tests {
		if got := deviceTypeFromMDNSService(serviceType); got != want {
			t.Fatalf("deviceTypeFromMDNSService(%q) = %q, want %q", serviceType, got, want)
		}
	}
}

func TestParseSSDPResponse(t *testing.T) {
	payload := "HTTP/1.1 200 OK\r\n" +
		"CACHE-CONTROL: max-age=1800\r\n" +
		"LOCATION: http://192.168.1.15:8008/ssdp/device-desc.xml\r\n" +
		"SERVER: Linux/5.10 UPnP/1.0 GoogleCast/1.0\r\n" +
		"ST: urn:dial-multiscreen-org:service:dial:1\r\n" +
		"USN: uuid:device-id::urn:dial-multiscreen-org:service:dial:1\r\n\r\n"

	response := parseSSDPResponse(payload)

	if got := ssdpHeader(response.Headers, "location"); got != "http://192.168.1.15:8008/ssdp/device-desc.xml" {
		t.Fatalf("location = %q, want descriptor URL", got)
	}
	if got := ssdpHeader(response.Headers, "server"); got != "Linux/5.10 UPnP/1.0 GoogleCast/1.0" {
		t.Fatalf("server = %q, want GoogleCast server header", got)
	}
	if got := deviceTypeFromSSDP(response.Headers); got != DeviceTV {
		t.Fatalf("deviceTypeFromSSDP() = %q, want %q", got, DeviceTV)
	}
}

func TestDeviceTypeFromSSDP(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    DeviceType
	}{
		{
			name: "router",
			headers: map[string]string{
				"st": "urn:schemas-upnp-org:device:InternetGatewayDevice:1",
			},
			want: DeviceRouter,
		},
		{
			name: "media renderer",
			headers: map[string]string{
				"st": "urn:schemas-upnp-org:device:MediaRenderer:1",
			},
			want: DeviceTV,
		},
		{
			name: "printer",
			headers: map[string]string{
				"server": "Printer UPnP/1.0",
			},
			want: DevicePrinter,
		},
		{
			name: "generic upnp",
			headers: map[string]string{
				"st": "upnp:rootdevice",
			},
			want: DeviceIoT,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := deviceTypeFromSSDP(test.headers); got != test.want {
				t.Fatalf("deviceTypeFromSSDP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSSDPHints(t *testing.T) {
	hints := ssdpHints(map[string]string{
		"st":       "upnp:rootdevice",
		"server":   "Linux/5.10 UPnP/1.0",
		"location": "http://192.168.1.20:1900/device.xml",
	})

	if len(hints) != 4 {
		t.Fatalf("ssdpHints() returned %d hints, want 4: %v", len(hints), hints)
	}
	if hints[0] != "Dispositivo anunciado por SSDP/UPnP." {
		t.Fatalf("first hint = %q, want SSDP discovery hint", hints[0])
	}
}

func TestCompactSSDPHints(t *testing.T) {
	hints := compactSSDPHints([]string{
		"Dispositivo anunciado por SSDP/UPnP.",
		"Servicio UPnP: service-1.",
		"Servicio UPnP: service-2.",
		"Servicio UPnP: service-3.",
		"Servicio UPnP: service-4.",
		"Servicio UPnP: service-5.",
		"Servicio UPnP: service-6.",
		"Servicio UPnP: service-7.",
		"Servidor UPnP: server-1.",
		"Servidor UPnP: server-2.",
		"Servidor UPnP: server-3.",
		"Descriptor UPnP disponible en http://192.168.1.2/a.xml.",
		"Descriptor UPnP disponible en http://192.168.1.2/b.xml.",
		"Descriptor UPnP disponible en http://192.168.1.2/c.xml.",
	})

	if len(hints) != 11 {
		t.Fatalf("compactSSDPHints() returned %d hints, want 11: %v", len(hints), hints)
	}
	if hints[len(hints)-1] != "Descriptor UPnP disponible en http://192.168.1.2/b.xml." {
		t.Fatalf("last hint = %q, want second descriptor", hints[len(hints)-1])
	}
}

func TestAppendUnique(t *testing.T) {
	got := appendUnique([]string{"one", "two"}, "two", "", "three")
	want := []string{"one", "two", "three"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("appendUnique() = %v, want %v", got, want)
	}
}
