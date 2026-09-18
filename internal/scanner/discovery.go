package scanner

import (
	"context"
	"fmt"

	"github.com/lcdosguzman/lansweepgo/internal/network"
)

type DiscoverySource string

const (
	DiscoverySourceTCP  DiscoverySource = "tcp"
	DiscoverySourceARP  DiscoverySource = "arp"
	DiscoverySourceMDNS DiscoverySource = "mdns"
	DiscoverySourceSSDP DiscoverySource = "ssdp"
)

type Discoverer interface {
	Source() DiscoverySource
	Discover(ctx context.Context, local network.LocalNetwork) ([]Device, error)
}

func defaultDiscoverers(config Config) []Discoverer {
	return []Discoverer{
		NewTCPDiscoverer(config),
		ARPDiscoverer{},
		MDNSDiscoverer{},
		SSDPDiscoverer{},
	}
}

func shouldEmitDiscoveryEvent(source DiscoverySource, alreadyKnown bool) bool {
	return source == DiscoverySourceTCP || !alreadyKnown
}

func discoveryEventMessage(source DiscoverySource, ip string) string {
	switch source {
	case DiscoverySourceARP:
		return fmt.Sprintf("Discovered %s from ARP cache", ip)
	case DiscoverySourceMDNS:
		return fmt.Sprintf("Discovered %s from mDNS/Bonjour", ip)
	case DiscoverySourceSSDP:
		return fmt.Sprintf("Discovered %s from SSDP/UPnP", ip)
	default:
		return fmt.Sprintf("Discovered %s", ip)
	}
}
