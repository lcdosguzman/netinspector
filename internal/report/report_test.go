package report

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lcdosguzman/netinspector/internal/scanner"
)

func TestJSONExportsFullScan(t *testing.T) {
	scan := sampleScan()

	payload, err := JSON(scan)
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var decoded scanner.ScanResult
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("exported JSON is invalid: %v", err)
	}
	if decoded.ID != scan.ID || len(decoded.Devices) != 1 {
		t.Fatalf("decoded scan = %+v, want original scan data", decoded)
	}
}

func TestCSVExportsDeviceRows(t *testing.T) {
	payload, err := CSV(sampleScan())
	if err != nil {
		t.Fatalf("CSV returned error: %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(string(payload))).ReadAll()
	if err != nil {
		t.Fatalf("exported CSV is invalid: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("CSV rows = %d, want header plus one device", len(rows))
	}
	if rows[0][0] != "scan_id" || rows[0][5] != "device_ip" {
		t.Fatalf("CSV header = %v, want scan and device columns", rows[0])
	}
	if rows[1][0] != "scan-1" || rows[1][5] != "192.168.1.7" {
		t.Fatalf("CSV device row = %v, want exported scan and device data", rows[1])
	}
	if rows[1][12] != "tcp/8443 https-alt" {
		t.Fatalf("CSV ports = %q, want formatted ports", rows[1][12])
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
