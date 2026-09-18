package demo

import (
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/scanner"
)

func NewScan() scanner.ScanResult {
	now := time.Now().UTC()
	devices := []scanner.Device{
		{
			IP:        "192.168.1.1",
			Hostname:  "home-gateway.local",
			MAC:       "40:9B:CD:10:2F:01",
			Vendor:    "Ubiquiti",
			Type:      scanner.DeviceRouter,
			LatencyMS: 2,
			IsActive:  true,
			Ports: []scanner.Port{
				{Number: 53, Protocol: "tcp", ServiceName: "dns"},
				{Number: 80, Protocol: "tcp", ServiceName: "http"},
				{Number: 443, Protocol: "tcp", ServiceName: "https"},
			},
		},
		{
			IP:        "192.168.1.24",
			Hostname:  "simon-macbook.local",
			MAC:       "A4:83:E7:91:AA:20",
			Vendor:    "Apple",
			Type:      scanner.DeviceDesktop,
			LatencyMS: 6,
			IsActive:  true,
			Ports: []scanner.Port{
				{Number: 22, Protocol: "tcp", ServiceName: "ssh"},
				{Number: 5900, Protocol: "tcp", ServiceName: "vnc"},
			},
		},
		{
			IP:        "192.168.1.36",
			Hostname:  "living-room-tv.local",
			MAC:       "F8:4E:73:22:18:99",
			Vendor:    "Samsung",
			Type:      scanner.DeviceTV,
			LatencyMS: 18,
			IsActive:  true,
			Ports: []scanner.Port{
				{Number: 8000, Protocol: "tcp", ServiceName: "http-alt"},
			},
		},
		{
			IP:        "192.168.1.48",
			Hostname:  "office-printer.local",
			MAC:       "30:05:5C:44:82:B1",
			Vendor:    "HP",
			Type:      scanner.DevicePrinter,
			LatencyMS: 12,
			IsActive:  true,
			Ports: []scanner.Port{
				{Number: 631, Protocol: "tcp", ServiceName: "ipp"},
			},
		},
	}

	events := []scanner.ScanEvent{{
		Type:      "SCAN_STARTED",
		Message:   "Demo scan started",
		Timestamp: now,
	}}
	for _, device := range devices {
		events = append(events, scanner.ScanEvent{
			Type:      "DEVICE_DISCOVERED",
			Message:   "Discovered " + device.Hostname,
			DeviceIP:  device.IP,
			Timestamp: now.Add(time.Duration(len(events)) * 250 * time.Millisecond),
		})
	}
	events = append(events, scanner.ScanEvent{
		Type:      "SCAN_FINISHED",
		Message:   "Demo scan finished",
		Timestamp: now.Add(2 * time.Second),
	})

	return scanner.ScanResult{
		ID:        "demo-scan",
		Mode:      "DEMO",
		Network:   "192.168.1.0/24",
		StartedAt: now,
		EndedAt:   now.Add(2 * time.Second),
		Devices:   devices,
		Events:    events,
	}
}
