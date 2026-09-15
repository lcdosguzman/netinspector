package service

import (
	"context"
	"errors"
	"testing"
)

func TestStartScanReturnsDemoScanByDefault(t *testing.T) {
	service := NewScanService(ScanServiceConfig{})

	result, err := service.StartScan(context.Background(), StartScanRequest{})
	if err != nil {
		t.Fatalf("StartScan returned error: %v", err)
	}

	if result.Mode != ScanModeDemo {
		t.Fatalf("Mode = %q, want %q", result.Mode, ScanModeDemo)
	}
	if len(result.Devices) == 0 {
		t.Fatal("expected demo devices")
	}
}

func TestStartScanRejectsUnsupportedMode(t *testing.T) {
	service := NewScanService(ScanServiceConfig{})

	_, err := service.StartScan(context.Background(), StartScanRequest{Mode: "INVALID"})
	if !errors.Is(err, ErrUnsupportedScanMode) {
		t.Fatalf("StartScan error = %v, want ErrUnsupportedScanMode", err)
	}
}
