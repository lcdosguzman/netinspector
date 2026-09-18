package storage

import (
	"context"
	"errors"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/scanner"
)

var ErrScanNotFound = errors.New("scan not found")

type ScanSummary struct {
	ID          string    `json:"id"`
	Mode        string    `json:"mode"`
	Network     string    `json:"network"`
	StartedAt   time.Time `json:"startedAt"`
	EndedAt     time.Time `json:"endedAt"`
	DeviceCount int       `json:"deviceCount"`
}

type ListScansFilter struct {
	Query string
	Limit int
}

type ScanRepository interface {
	SaveScan(ctx context.Context, result scanner.ScanResult) error
	ListScans(ctx context.Context, filter ListScansFilter) ([]ScanSummary, error)
	GetScan(ctx context.Context, id string) (scanner.ScanResult, error)
	Close() error
}
