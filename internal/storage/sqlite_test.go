package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/scanner"
)

func TestWhenSavingScanThenRepositoryListsAndGetsIt(t *testing.T) {
	cases := []struct {
		name       string
		result     scanner.ScanResult
		query      string
		wantCount  int
		wantHost   string
		wantDevice int
	}{
		{
			name: "when scan is saved then it can be found by device vendor",
			result: scanner.ScanResult{
				ID:        "scan-test",
				Mode:      "REAL",
				Network:   "192.168.1.0/24",
				StartedAt: time.Unix(100, 0).UTC(),
				EndedAt:   time.Unix(101, 0).UTC(),
				Devices: []scanner.Device{{
					IP:       "192.168.1.10",
					Hostname: "living-room-tv.local",
					Vendor:   "Samsung",
					Type:     scanner.DeviceTV,
					IsActive: true,
				}},
				Events: []scanner.ScanEvent{{
					Type:      "SCAN_FINISHED",
					Message:   "Scan finished with 1 active devices",
					Timestamp: time.Unix(101, 0).UTC(),
				}},
			},
			query:      "Samsung",
			wantCount:  1,
			wantHost:   "living-room-tv.local",
			wantDevice: 1,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository, err := NewSQLiteScanRepository(filepath.Join(t.TempDir(), "history.db"))
			if err != nil {
				t.Fatalf("NewSQLiteScanRepository returned error: %v", err)
			}
			defer repository.Close()

			if err := repository.SaveScan(context.Background(), test.result); err != nil {
				t.Fatalf("SaveScan returned error: %v", err)
			}

			scans, err := repository.ListScans(context.Background(), ListScansFilter{Query: test.query})
			if err != nil {
				t.Fatalf("ListScans returned error: %v", err)
			}
			if len(scans) != test.wantCount {
				t.Fatalf("ListScans returned %d scans, want %d", len(scans), test.wantCount)
			}
			if scans[0].DeviceCount != test.wantDevice {
				t.Fatalf("DeviceCount = %d, want %d", scans[0].DeviceCount, test.wantDevice)
			}

			stored, err := repository.GetScan(context.Background(), test.result.ID)
			if err != nil {
				t.Fatalf("GetScan returned error: %v", err)
			}
			if stored.Devices[0].Hostname != test.wantHost {
				t.Fatalf("stored hostname = %q, want %q", stored.Devices[0].Hostname, test.wantHost)
			}
		})
	}
}

func TestWhenGettingMissingScanThenNotFoundIsReturned(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{name: "when scan id does not exist then not found error is returned", id: "missing"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository, err := NewSQLiteScanRepository(filepath.Join(t.TempDir(), "history.db"))
			if err != nil {
				t.Fatalf("NewSQLiteScanRepository returned error: %v", err)
			}
			defer repository.Close()

			_, err = repository.GetScan(context.Background(), test.id)
			if err != ErrScanNotFound {
				t.Fatalf("GetScan error = %v, want ErrScanNotFound", err)
			}
		})
	}
}
