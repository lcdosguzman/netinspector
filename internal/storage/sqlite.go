package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/scanner"
	_ "modernc.org/sqlite"
)

const defaultScanHistoryLimit = 25

type SQLiteScanRepository struct {
	db *sql.DB
}

func NewSQLiteScanRepository(path string) (*SQLiteScanRepository, error) {
	if path == "" {
		path = filepath.Join(".lansweepgo", "lansweepgo.db")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	repository := &SQLiteScanRepository{db: db}
	if err := repository.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return repository, nil
}

func (repository *SQLiteScanRepository) Close() error {
	if repository == nil || repository.db == nil {
		return nil
	}
	return repository.db.Close()
}

func (repository *SQLiteScanRepository) SaveScan(ctx context.Context, result scanner.ScanResult) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal scan result: %w", err)
	}

	_, err = repository.db.ExecContext(ctx, `
		INSERT INTO scans (
			id,
			mode,
			network,
			started_at,
			ended_at,
			device_count,
			result_json
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			mode = excluded.mode,
			network = excluded.network,
			started_at = excluded.started_at,
			ended_at = excluded.ended_at,
			device_count = excluded.device_count,
			result_json = excluded.result_json
	`, result.ID, result.Mode, result.Network, result.StartedAt.UnixMilli(), result.EndedAt.UnixMilli(), len(result.Devices), string(payload))
	if err != nil {
		return fmt.Errorf("save scan: %w", err)
	}

	return nil
}

func (repository *SQLiteScanRepository) ListScans(ctx context.Context, filter ListScansFilter) ([]ScanSummary, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultScanHistoryLimit
	}
	if limit > 100 {
		limit = 100
	}

	query := strings.TrimSpace(filter.Query)
	args := []any{}
	statement := `
		SELECT id, mode, network, started_at, ended_at, device_count
		FROM scans
	`
	if query != "" {
		statement += `
			WHERE id LIKE ?
				OR mode LIKE ?
				OR network LIKE ?
				OR result_json LIKE ?
		`
		likeQuery := "%" + query + "%"
		args = append(args, likeQuery, likeQuery, likeQuery, likeQuery)
	}
	statement += `
		ORDER BY started_at DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := repository.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, fmt.Errorf("list scans: %w", err)
	}
	defer rows.Close()

	scans := make([]ScanSummary, 0)
	for rows.Next() {
		var summary ScanSummary
		var startedAt int64
		var endedAt int64
		if err := rows.Scan(&summary.ID, &summary.Mode, &summary.Network, &startedAt, &endedAt, &summary.DeviceCount); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		summary.StartedAt = time.UnixMilli(startedAt).UTC()
		summary.EndedAt = time.UnixMilli(endedAt).UTC()
		scans = append(scans, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scan history: %w", err)
	}

	return scans, nil
}

func (repository *SQLiteScanRepository) GetScan(ctx context.Context, id string) (scanner.ScanResult, error) {
	var payload string
	err := repository.db.QueryRowContext(ctx, `
		SELECT result_json
		FROM scans
		WHERE id = ?
	`, id).Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return scanner.ScanResult{}, ErrScanNotFound
		}
		return scanner.ScanResult{}, fmt.Errorf("get scan: %w", err)
	}

	var result scanner.ScanResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return scanner.ScanResult{}, fmt.Errorf("unmarshal scan result: %w", err)
	}

	return result, nil
}

func (repository *SQLiteScanRepository) migrate(ctx context.Context) error {
	_, err := repository.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS scans (
			id TEXT PRIMARY KEY,
			mode TEXT NOT NULL,
			network TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			ended_at INTEGER NOT NULL,
			device_count INTEGER NOT NULL,
			result_json TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS scans_started_at_idx ON scans(started_at DESC);
		CREATE INDEX IF NOT EXISTS scans_network_idx ON scans(network);
	`)
	if err != nil {
		return fmt.Errorf("migrate sqlite database: %w", err)
	}
	return nil
}
