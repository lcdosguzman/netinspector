package scanner

import (
	"reflect"
	"testing"
)

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
