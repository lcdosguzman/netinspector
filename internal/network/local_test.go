package network

import (
	"net"
	"testing"
)

func TestWhenCheckingInterfaceUsabilityThenOnlyActiveNonLoopbackInterfacesPass(t *testing.T) {
	cases := []struct {
		name  string
		iface net.Interface
		want  bool
	}{
		{
			name:  "when interface is up and not loopback then it is usable",
			iface: net.Interface{Flags: net.FlagUp},
			want:  true,
		},
		{
			name:  "when interface is down then it is not usable",
			iface: net.Interface{Flags: 0},
			want:  false,
		},
		{
			name:  "when interface is loopback then it is not usable",
			iface: net.Interface{Flags: net.FlagUp | net.FlagLoopback},
			want:  false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := isUsableInterface(test.iface); got != test.want {
				t.Fatalf("isUsableInterface() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWhenExtractingIPv4NetworkThenOnlyIPv4AddressesAreAccepted(t *testing.T) {
	cases := []struct {
		name        string
		address     net.Addr
		wantOK      bool
		wantIP      string
		wantNetwork string
	}{
		{
			name: "when address is IPv4 then network is returned",
			address: &net.IPNet{
				IP:   net.ParseIP("192.168.1.42"),
				Mask: net.CIDRMask(24, 32),
			},
			wantOK:      true,
			wantIP:      "192.168.1.42",
			wantNetwork: "192.168.1.0/24",
		},
		{
			name: "when address is IPv6 then it is rejected",
			address: &net.IPNet{
				IP:   net.ParseIP("fe80::1"),
				Mask: net.CIDRMask(64, 128),
			},
			wantOK: false,
		},
		{
			name:    "when address is not an IP network then it is rejected",
			address: &net.UnixAddr{Name: "socket", Net: "unix"},
			wantOK:  false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			ip, network, ok := ipv4Network(test.address)
			if ok != test.wantOK {
				t.Fatalf("ok = %v, want %v", ok, test.wantOK)
			}
			if !test.wantOK {
				return
			}
			if got := ip.String(); got != test.wantIP {
				t.Fatalf("ip = %q, want %q", got, test.wantIP)
			}
			if got := network.String(); got != test.wantNetwork {
				t.Fatalf("network = %q, want %q", got, test.wantNetwork)
			}
		})
	}
}
