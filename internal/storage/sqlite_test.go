package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/lcdosguzman/netinspector/internal/scanner"
)

func TestSQLiteScanRepositorySavesListsAndGetsScans(t *testing.T) {
	repository, err := NewSQLiteScanRepository(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatalf("NewSQLiteScanRepository returned error: %v", err)
	}
	defer repository.Close()

	result := scanner.ScanResult{
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
	}

	if err := repository.SaveScan(context.Background(), result); err != nil {
		t.Fatalf("SaveScan returned error: %v", err)
	}

	scans, err := repository.ListScans(context.Background(), ListScansFilter{Query: "Samsung"})
	if err != nil {
		t.Fatalf("ListScans returned error: %v", err)
	}
	if len(scans) != 1 {
		t.Fatalf("ListScans returned %d scans, want 1", len(scans))
	}
	if scans[0].DeviceCount != 1 {
		t.Fatalf("DeviceCount = %d, want 1", scans[0].DeviceCount)
	}

	stored, err := repository.GetScan(context.Background(), "scan-test")
	if err != nil {
		t.Fatalf("GetScan returned error: %v", err)
	}
	if stored.Devices[0].Hostname != "living-room-tv.local" {
		t.Fatalf("stored hostname = %q, want living-room-tv.local", stored.Devices[0].Hostname)
	}
}

func TestSQLiteScanRepositoryReturnsNotFound(t *testing.T) {
	repository, err := NewSQLiteScanRepository(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatalf("NewSQLiteScanRepository returned error: %v", err)
	}
	defer repository.Close()

	_, err = repository.GetScan(context.Background(), "missing")
	if err != ErrScanNotFound {
		t.Fatalf("GetScan error = %v, want ErrScanNotFound", err)
	}
}
