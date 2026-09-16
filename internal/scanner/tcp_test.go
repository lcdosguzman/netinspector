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

func TestWhenBuildingHostsThenNetworkAndBroadcastAreHandledByCIDRSize(t *testing.T) {
	cases := []struct {
		name string
		cidr string
		want []string
	}{
		{
			name: "when network is point to point then both addresses are kept",
			cidr: "192.168.1.0/31",
			want: []string{"192.168.1.0", "192.168.1.1"},
		},
		{
			name: "when network has broadcast address then network and broadcast are skipped",
			cidr: "192.168.1.0/30",
			want: []string{"192.168.1.1", "192.168.1.2"},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := hosts(test.cidr)
			if err != nil {
				t.Fatalf("hosts returned error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("hosts() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWhenBuildingHostsThenInvalidCIDRsAreRejected(t *testing.T) {
	cases := []struct {
		name string
		cidr string
	}{
		{name: "when CIDR syntax is invalid then error is returned", cidr: "not-a-cidr"},
		{name: "when CIDR is IPv6 then error is returned", cidr: "fe80::/64"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := hosts(test.cidr); err == nil {
				t.Fatal("hosts returned nil error, want failure")
			}
		})
	}
}

func TestWhenScanningCIDRThenConfiguredDiscoverersAreUsed(t *testing.T) {
	cases := []struct {
		name          string
		discoverers   []Discoverer
		wantDevices   int
		wantEventText []string
	}{
		{
			name: "when TCP and ARP discoverers return devices then scan merges them",
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
			wantDevices: 2,
			wantEventText: []string{
				"Discovered 192.168.1.2",
				"Discovered 192.168.1.3 from ARP cache",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			scanner := TCPScanner{
				config:      normalizeConfig(Config{}),
				discoverers: test.discoverers,
			}

			result, err := scanner.ScanCIDR(context.Background(), network.LocalNetwork{CIDR: "192.168.1.0/30"})
			if err != nil {
				t.Fatalf("ScanCIDR returned error: %v", err)
			}

			if len(result.Devices) != test.wantDevices {
				t.Fatalf("ScanCIDR returned %d devices, want %d", len(result.Devices), test.wantDevices)
			}
			for index, want := range test.wantEventText {
				if got := result.Events[index+1].Message; got != want {
					t.Fatalf("event %d = %q, want %q", index+1, got, want)
				}
			}
		})
	}
}

func TestWhenResolvingServiceNameThenKnownPortsUseFriendlyNames(t *testing.T) {
	cases := []struct {
		name string
		port int
		want string
	}{
		{name: "when port is SSH then service name is ssh", port: 22, want: "ssh"},
		{name: "when port is HTTPS then service name is https", port: 443, want: "https"},
		{name: "when port is unknown then service name is unknown", port: 65535, want: "unknown"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := serviceName(test.port); got != test.want {
				t.Fatalf("serviceName(%d) = %q, want %q", test.port, got, test.want)
			}
		})
	}
}

func TestWhenFormattingDiscoveryEventThenSourceSpecificMessageIsUsed(t *testing.T) {
	cases := []struct {
		name   string
		source DiscoverySource
		want   string
	}{
		{name: "when source is TCP then default discovery message is used", source: DiscoverySourceTCP, want: "Discovered 192.168.1.9"},
		{name: "when source is ARP then ARP cache message is used", source: DiscoverySourceARP, want: "Discovered 192.168.1.9 from ARP cache"},
		{name: "when source is mDNS then Bonjour message is used", source: DiscoverySourceMDNS, want: "Discovered 192.168.1.9 from mDNS/Bonjour"},
		{name: "when source is SSDP then UPnP message is used", source: DiscoverySourceSSDP, want: "Discovered 192.168.1.9 from SSDP/UPnP"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := discoveryEventMessage(test.source, "192.168.1.9"); got != test.want {
				t.Fatalf("discoveryEventMessage() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWhenMergingDeviceThenExistingDataIsPreservedAndNewEvidenceIsAdded(t *testing.T) {
	cases := []struct {
		name      string
		existing  []Device
		next      Device
		wantCount int
		assert    func(t *testing.T, device Device)
	}{
		{
			name: "when same IP is discovered again then identity evidence is merged",
			existing: []Device{{
				IP:        "192.168.1.7",
				Hostname:  "tv.local",
				Type:      DeviceUnknown,
				Hints:     []string{"existing hint"},
				IsActive:  true,
				LatencyMS: 10,
				Ports:     []Port{{Number: 8008, Protocol: "tcp", ServiceName: "http-alt"}},
			}},
			next: Device{
				IP:       "192.168.1.7",
				MAC:      "78:66:9d:0b:bc:16",
				Vendor:   "Hui Zhou Gaoshengda Technology Co., Ltd.",
				Type:     DeviceTV,
				Hints:    []string{"existing hint", "new hint"},
				IsActive: true,
				Ports:    []Port{{Number: 8443, Protocol: "tcp", ServiceName: "https-alt"}},
			},
			wantCount: 1,
			assert: func(t *testing.T, device Device) {
				t.Helper()
				if device.Hostname != "tv.local" || device.Type != DeviceTV || device.MAC == "" {
					t.Fatalf("merged device = %+v, want preserved hostname plus new identity", device)
				}
				if len(device.Hints) != 2 {
					t.Fatalf("hints = %v, want unique existing and new hints", device.Hints)
				}
				if len(device.Ports) != 2 {
					t.Fatalf("ports = %v, want merged unique ports", device.Ports)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			merged := mergeDevice(test.existing, test.next)
			if len(merged) != test.wantCount {
				t.Fatalf("merged length = %d, want %d", len(merged), test.wantCount)
			}
			test.assert(t, merged[0])
		})
	}
}

func TestWhenParsingARPNeighborsThenInvalidEntriesAreFiltered(t *testing.T) {
	cases := []struct {
		name      string
		output    string
		cidr      string
		wantCount int
		assert    func(t *testing.T, devices []Device)
	}{
		{
			name: "when ARP output contains incomplete broadcast and multicast rows then only valid neighbors remain",
			output: `? (192.168.1.1) at 84:15:d3:1b:8e:dd on en0 ifscope [ethernet]
? (192.168.1.7) at 78:66:9d:b:bc:16 on en0 ifscope [ethernet]
? (192.168.1.8) at (incomplete) on en0 ifscope [ethernet]
? (192.168.1.255) at ff:ff:ff:ff:ff:ff on en0 ifscope [ethernet]
? (224.0.0.251) at 1:0:5e:0:0:fb on en0 ifscope permanent [ethernet]`,
			cidr:      "192.168.1.0/24",
			wantCount: 2,
			assert: func(t *testing.T, devices []Device) {
				t.Helper()
				if devices[0].IP != "192.168.1.1" || devices[0].Type != DeviceRouter {
					t.Fatalf("first ARP neighbor = %+v, want router 192.168.1.1", devices[0])
				}
				if devices[1].IP != "192.168.1.7" || devices[1].MAC != "78:66:9d:b:bc:16" {
					t.Fatalf("second ARP neighbor = %+v, want 192.168.1.7 with MAC", devices[1])
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := parseARPNeighbors(test.output, test.cidr)
			if len(got) != test.wantCount {
				t.Fatalf("parseARPNeighbors() returned %d devices, want %d", len(got), test.wantCount)
			}
			test.assert(t, got)
		})
	}
}

func TestWhenResolvingVendorThenLocallyAdministeredMACIsMarkedPrivate(t *testing.T) {
	cases := []struct {
		name     string
		mac      string
		fallback string
		want     string
	}{
		{
			name: "when MAC is locally administered then vendor is private randomized marker",
			mac:  "9e:68:c4:89:82:06",
			want: "Private/randomized MAC",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := vendorName(test.mac, test.fallback); got != test.want {
				t.Fatalf("vendorName() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWhenEnrichingDeviceThenOUIAndPortsInferIdentity(t *testing.T) {
	cases := []struct {
		name       string
		device     Device
		wantVendor string
		wantType   DeviceType
	}{
		{
			name: "when OUI and media port match TV patterns then device is TV",
			device: Device{
				IP:       "192.168.1.7",
				MAC:      "78:66:9d:0b:bc:16",
				Type:     DeviceUnknown,
				IsActive: true,
				Ports: []Port{
					{Number: 8443, Protocol: "tcp", ServiceName: "https-alt"},
				},
			},
			wantVendor: "Hui Zhou Gaoshengda Technology Co., Ltd.",
			wantType:   DeviceTV,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			device := enrichDevice(test.device)

			if device.Vendor != test.wantVendor {
				t.Fatalf("Vendor = %q, want %q", device.Vendor, test.wantVendor)
			}
			if device.Type != test.wantType {
				t.Fatalf("Type = %q, want %q", device.Type, test.wantType)
			}
			if len(device.Hints) == 0 {
				t.Fatal("expected identification hints")
			}
		})
	}
}

func TestWhenEnrichingDeviceThenComputerHostnameOverridesAirPlayHint(t *testing.T) {
	cases := []struct {
		name     string
		device   Device
		wantType DeviceType
	}{
		{
			name: "when hostname is MacBook then desktop type overrides AirPlay TV hint",
			device: Device{
				IP:       "192.168.1.11",
				MAC:      "3c:06:30:4c:e0:d1",
				Hostname: "Simons-MacBook-Pro.local",
				Type:     DeviceTV,
				IsActive: true,
				Hints: []string{
					"Servicio multimedia anunciado por Bonjour; posible TV, speaker o dispositivo cast.",
				},
			},
			wantType: DeviceDesktop,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			device := enrichDevice(test.device)
			if device.Type != test.wantType {
				t.Fatalf("Type = %q, want %q", device.Type, test.wantType)
			}
		})
	}
}

func TestWhenInferringTypeFromMDNSServiceThenKnownServicesMapToDeviceTypes(t *testing.T) {
	cases := []struct {
		name        string
		serviceType string
		want        DeviceType
	}{
		{name: "when service is IPP then device is printer", serviceType: "_ipp._tcp", want: DevicePrinter},
		{name: "when service is AirPlay then device is TV", serviceType: "_airplay._tcp", want: DeviceTV},
		{name: "when service is Google Cast then device is TV", serviceType: "_googlecast._tcp", want: DeviceTV},
		{name: "when service is workstation then device is desktop", serviceType: "_workstation._tcp", want: DeviceDesktop},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := deviceTypeFromMDNSService(test.serviceType); got != test.want {
				t.Fatalf("deviceTypeFromMDNSService(%q) = %q, want %q", test.serviceType, got, test.want)
			}
		})
	}
}

func TestWhenParsingSSDPResponseThenHeadersAndDeviceTypeAreExtracted(t *testing.T) {
	cases := []struct {
		name         string
		payload      string
		wantLocation string
		wantServer   string
		wantType     DeviceType
	}{
		{
			name: "when response is Google Cast DIAL then descriptor and TV type are extracted",
			payload: "HTTP/1.1 200 OK\r\n" +
				"CACHE-CONTROL: max-age=1800\r\n" +
				"LOCATION: http://192.168.1.15:8008/ssdp/device-desc.xml\r\n" +
				"SERVER: Linux/5.10 UPnP/1.0 GoogleCast/1.0\r\n" +
				"ST: urn:dial-multiscreen-org:service:dial:1\r\n" +
				"USN: uuid:device-id::urn:dial-multiscreen-org:service:dial:1\r\n\r\n",
			wantLocation: "http://192.168.1.15:8008/ssdp/device-desc.xml",
			wantServer:   "Linux/5.10 UPnP/1.0 GoogleCast/1.0",
			wantType:     DeviceTV,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			response := parseSSDPResponse(test.payload)

			if got := ssdpHeader(response.Headers, "location"); got != test.wantLocation {
				t.Fatalf("location = %q, want %q", got, test.wantLocation)
			}
			if got := ssdpHeader(response.Headers, "server"); got != test.wantServer {
				t.Fatalf("server = %q, want %q", got, test.wantServer)
			}
			if got := deviceTypeFromSSDP(response.Headers); got != test.wantType {
				t.Fatalf("deviceTypeFromSSDP() = %q, want %q", got, test.wantType)
			}
		})
	}
}

func TestWhenInferringTypeFromSSDPThenKnownUPnPSignalsMapToDeviceTypes(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		want    DeviceType
	}{
		{
			name: "when SSDP advertises internet gateway then device is router",
			headers: map[string]string{
				"st": "urn:schemas-upnp-org:device:InternetGatewayDevice:1",
			},
			want: DeviceRouter,
		},
		{
			name: "when SSDP advertises media renderer then device is TV",
			headers: map[string]string{
				"st": "urn:schemas-upnp-org:device:MediaRenderer:1",
			},
			want: DeviceTV,
		},
		{
			name: "when SSDP server mentions printer then device is printer",
			headers: map[string]string{
				"server": "Printer UPnP/1.0",
			},
			want: DevicePrinter,
		},
		{
			name: "when SSDP advertises generic UPnP then device is IoT",
			headers: map[string]string{
				"st": "upnp:rootdevice",
			},
			want: DeviceIoT,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := deviceTypeFromSSDP(test.headers); got != test.want {
				t.Fatalf("deviceTypeFromSSDP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWhenBuildingSSDPHintsThenAvailableHeadersBecomeEvidence(t *testing.T) {
	cases := []struct {
		name      string
		headers   map[string]string
		wantCount int
	}{
		{
			name: "when headers include service server and location then hints include each one",
			headers: map[string]string{
				"st":       "upnp:rootdevice",
				"server":   "Linux/5.10 UPnP/1.0",
				"location": "http://192.168.1.20:1900/device.xml",
			},
			wantCount: 4,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			hints := ssdpHints(test.headers)

			if len(hints) != test.wantCount {
				t.Fatalf("ssdpHints() returned %d hints, want %d: %v", len(hints), test.wantCount, hints)
			}
			if hints[0] != "Dispositivo anunciado por SSDP/UPnP." {
				t.Fatalf("first hint = %q, want SSDP discovery hint", hints[0])
			}
		})
	}
}

func TestWhenCompactingSSDPHintsThenRepeatedEvidenceIsLimited(t *testing.T) {
	cases := []struct {
		name     string
		hints    []string
		wantLen  int
		wantLast string
	}{
		{
			name: "when hints contain repeated services servers and descriptors then only allowed counts remain",
			hints: []string{
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
			},
			wantLen:  11,
			wantLast: "Descriptor UPnP disponible en http://192.168.1.2/b.xml.",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			hints := compactSSDPHints(test.hints)

			if len(hints) != test.wantLen {
				t.Fatalf("compactSSDPHints() returned %d hints, want %d: %v", len(hints), test.wantLen, hints)
			}
			if hints[len(hints)-1] != test.wantLast {
				t.Fatalf("last hint = %q, want %q", hints[len(hints)-1], test.wantLast)
			}
		})
	}
}

func TestWhenAppendingHintsThenBlankAndDuplicateValuesAreSkipped(t *testing.T) {
	cases := []struct {
		name     string
		existing []string
		next     []string
		want     []string
	}{
		{
			name:     "when values include blanks and duplicates then only unique nonblank values remain",
			existing: []string{"one", "two"},
			next:     []string{"two", "", "three"},
			want:     []string{"one", "two", "three"},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := appendUnique(test.existing, test.next...)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("appendUnique() = %v, want %v", got, test.want)
			}
		})
	}
}
