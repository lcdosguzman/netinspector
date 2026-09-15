package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lcdosguzman/netinspector/internal/demo"
	"github.com/lcdosguzman/netinspector/internal/network"
	"github.com/lcdosguzman/netinspector/internal/scanner"
	"github.com/lcdosguzman/netinspector/internal/storage"
)

const (
	ScanModeDemo = "DEMO"
	ScanModeReal = "REAL"
)

var (
	ErrUnsupportedScanMode     = errors.New("unsupported scan mode")
	ErrLocalNetworkUnavailable = errors.New("local network unavailable")
)

type ScanService struct {
	repository  storage.ScanRepository
	scanTimeout time.Duration
	concurrency int
}

type ScanServiceConfig struct {
	Repository  storage.ScanRepository
	ScanTimeout time.Duration
	Concurrency int
}

type StartScanRequest struct {
	Mode string
	CIDR string
}

func NewScanService(config ScanServiceConfig) ScanService {
	if config.ScanTimeout <= 0 {
		config.ScanTimeout = 300 * time.Millisecond
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 128
	}

	return ScanService{
		repository:  config.Repository,
		scanTimeout: config.ScanTimeout,
		concurrency: config.Concurrency,
	}
}

func (service ScanService) StartScan(ctx context.Context, request StartScanRequest) (scanner.ScanResult, error) {
	switch normalizeScanMode(request.Mode) {
	case "", ScanModeDemo:
		return demo.NewScan(), nil
	case ScanModeReal:
		return service.startRealScan(ctx, request.CIDR)
	default:
		return scanner.ScanResult{}, fmt.Errorf("%w %q", ErrUnsupportedScanMode, request.Mode)
	}
}

func (service ScanService) startRealScan(ctx context.Context, cidr string) (scanner.ScanResult, error) {
	local, err := network.DetectLocalNetwork()
	if err != nil {
		return scanner.ScanResult{}, fmt.Errorf("%w: %v", ErrLocalNetworkUnavailable, err)
	}
	if cidr != "" {
		local.CIDR = cidr
	}

	tcpScanner := scanner.NewTCPScanner(scanner.Config{
		Ports:       scanner.DefaultDiscoveryPorts(),
		Concurrency: service.concurrency,
		Timeout:     service.scanTimeout,
	})
	result, err := tcpScanner.ScanCIDR(ctx, local)
	if err != nil {
		return scanner.ScanResult{}, err
	}

	if service.repository != nil {
		if err := service.repository.SaveScan(ctx, result); err != nil {
			return scanner.ScanResult{}, err
		}
	}

	return result, nil
}

func normalizeScanMode(mode string) string {
	return strings.ToUpper(strings.TrimSpace(mode))
}
