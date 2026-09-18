package report

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/scanner"
)

func TestWhenExportingJSONThenScanPayloadIsPreserved(t *testing.T) {
	cases := []struct {
		name        string
		scan        scanner.ScanResult
		wantDevices int
	}{
		{name: "when scan has devices then JSON includes devices", scan: sampleScan(), wantDevices: 1},
		{name: "when scan has no devices then JSON includes empty device list", scan: scanner.ScanResult{ID: "scan-empty", Mode: "REAL"}, wantDevices: 0},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			payload, err := JSON(test.scan)
			if err != nil {
				t.Fatalf("JSON returned error: %v", err)
			}

			var decoded scanner.ScanResult
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatalf("exported JSON is invalid: %v", err)
			}
			if decoded.ID != test.scan.ID || len(decoded.Devices) != test.wantDevices {
				t.Fatalf("decoded scan = %+v, want original scan data", decoded)
			}
		})
	}
}

func TestWhenExportingCSVThenDeviceRowsAreWritten(t *testing.T) {
	cases := []struct {
		name     string
		scan     scanner.ScanResult
		wantRows int
	}{
		{name: "when scan has one device then CSV has header and device row", scan: sampleScan(), wantRows: 2},
		{name: "when scan has no devices then CSV has only header row", scan: scanner.ScanResult{ID: "scan-empty", Mode: "REAL"}, wantRows: 1},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			payload, err := CSV(test.scan)
			if err != nil {
				t.Fatalf("CSV returned error: %v", err)
			}

			rows, err := csv.NewReader(strings.NewReader(string(payload))).ReadAll()
			if err != nil {
				t.Fatalf("exported CSV is invalid: %v", err)
			}
			if len(rows) != test.wantRows {
				t.Fatalf("CSV rows = %d, want %d", len(rows), test.wantRows)
			}
			if rows[0][0] != "scan_id" || rows[0][5] != "device_ip" {
				t.Fatalf("CSV header = %v, want scan and device columns", rows[0])
			}
			if len(rows) == 1 {
				return
			}
			if rows[1][0] != "scan-1" || rows[1][5] != "192.168.1.7" {
				t.Fatalf("CSV device row = %v, want exported scan and device data", rows[1])
			}
			if rows[1][12] != "tcp/8443 https-alt" {
				t.Fatalf("CSV ports = %q, want formatted ports", rows[1][12])
			}
		})
	}
}

func TestWhenFormattingPortsThenPortsAreJoinedForReports(t *testing.T) {
	cases := []struct {
		name  string
		ports []scanner.Port
		want  string
	}{
		{name: "when ports are empty then formatted value is empty", ports: nil, want: ""},
		{
			name: "when one port is present then formatted value contains one entry",
			ports: []scanner.Port{{
				Number:      443,
				Protocol:    "tcp",
				ServiceName: "https",
			}},
			want: "tcp/443 https",
		},
		{
			name: "when multiple ports are present then formatted value joins entries",
			ports: []scanner.Port{
				{Number: 80, Protocol: "tcp", ServiceName: "http"},
				{Number: 8443, Protocol: "tcp", ServiceName: "https-alt"},
			},
			want: "tcp/80 http | tcp/8443 https-alt",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := formatPorts(test.ports); got != test.want {
				t.Fatalf("formatPorts() = %q, want %q", got, test.want)
			}
		})
	}
}

func sampleScan() scanner.ScanResult {
	startedAt := time.Date(2026, 9, 15, 10, 30, 0, 0, time.UTC)
	return scanner.ScanResult{
		ID:        "scan-1",
		Mode:      "REAL",
		Network:   "192.168.1.0/24",
		StartedAt: startedAt,
		EndedAt:   startedAt.Add(4 * time.Second),
		Devices: []scanner.Device{{
			IP:        "192.168.1.7",
			Hostname:  "living-room-tv.local",
			MAC:       "78:66:9d:0b:bc:16",
			Vendor:    "Hui Zhou Gaoshengda Technology Co., Ltd.",
			Type:      scanner.DeviceTV,
			Hints:     []string{"Hostname discovered", "SSDP media renderer"},
			LatencyMS: 23,
			IsActive:  true,
			Ports: []scanner.Port{{
				Number:      8443,
				Protocol:    "tcp",
				ServiceName: "https-alt",
			}},
		}},
	}
}
