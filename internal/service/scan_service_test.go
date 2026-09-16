package service

import (
	"context"
	"errors"
	"testing"
)

func TestWhenStartingScanInDemoModeThenDemoScanIsReturned(t *testing.T) {
	cases := []struct {
		name string
		mode string
	}{
		{name: "when mode is empty then demo scan is returned", mode: ""},
		{name: "when mode is explicit demo then demo scan is returned", mode: ScanModeDemo},
		{name: "when mode is lowercase demo then demo scan is returned", mode: "demo"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			service := NewScanService(ScanServiceConfig{})

			result, err := service.StartScan(context.Background(), StartScanRequest{Mode: test.mode})
			if err != nil {
				t.Fatalf("StartScan returned error: %v", err)
			}

			if result.Mode != ScanModeDemo {
				t.Fatalf("Mode = %q, want %q", result.Mode, ScanModeDemo)
			}
			if len(result.Devices) == 0 {
				t.Fatal("expected demo devices")
			}
		})
	}
}

func TestWhenStartingScanWithUnsupportedModeThenErrorIsReturned(t *testing.T) {
	cases := []struct {
		name string
		mode string
	}{
		{name: "when mode is unknown then unsupported mode error is returned", mode: "INVALID"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			service := NewScanService(ScanServiceConfig{})

			_, err := service.StartScan(context.Background(), StartScanRequest{Mode: test.mode})
			if !errors.Is(err, ErrUnsupportedScanMode) {
				t.Fatalf("StartScan error = %v, want ErrUnsupportedScanMode", err)
			}
		})
	}
}
