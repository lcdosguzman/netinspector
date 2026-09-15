package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	networkv1connect "github.com/lcdosguzman/netinspector/internal/gen/network/v1/networkv1connect"
	"github.com/lcdosguzman/netinspector/internal/network"
	"github.com/lcdosguzman/netinspector/internal/report"
	"github.com/lcdosguzman/netinspector/internal/scanner"
	appservice "github.com/lcdosguzman/netinspector/internal/service"
	"github.com/lcdosguzman/netinspector/internal/storage"
)

type Config struct {
	Addr         string
	ScanTimeout  time.Duration
	DatabasePath string
}

type Server struct {
	config     Config
	repository storage.ScanRepository
	scans      appservice.ScanService
}

func NewServer(config Config) Server {
	if config.Addr == "" {
		config.Addr = "127.0.0.1:8088"
	}
	if config.ScanTimeout <= 0 {
		config.ScanTimeout = 300 * time.Millisecond
	}
	if config.DatabasePath == "" {
		config.DatabasePath = os.Getenv("NETINSPECTOR_DB_PATH")
	}

	return Server{config: config}
}

func (server Server) ListenAndServe(ctx context.Context) error {
	repository, err := storage.NewSQLiteScanRepository(server.config.DatabasePath)
	if err != nil {
		return err
	}
	defer repository.Close()

	server.repository = repository
	server.scans = appservice.NewScanService(appservice.ScanServiceConfig{
		Repository:  repository,
		ScanTimeout: server.config.ScanTimeout,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /api/local-network", server.handleLocalNetwork)
	mux.HandleFunc("GET /api/scans", server.handleListScans)
	mux.HandleFunc("GET /api/scans/{id}/export/{format}", server.handleExportScan)
	mux.HandleFunc("GET /api/scans/{id}", server.handleGetScan)
	mux.HandleFunc("POST /api/scans", server.handleStartScan)
	path, handler := networkv1connect.NewNetworkServiceHandler(newNetworkService(repository, server.scans))
	mux.Handle(path, handler)

	httpServer := &http.Server{
		Addr:              server.config.Addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
			return
		}
		errs <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errs:
		return err
	}
}

func (server Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (server Server) handleLocalNetwork(w http.ResponseWriter, _ *http.Request) {
	local, err := network.DetectLocalNetwork()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	writeJSON(w, http.StatusOK, local)
}

func (server Server) handleStartScan(w http.ResponseWriter, r *http.Request) {
	var request startScanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request body: %w", err))
		return
	}

	result, err := server.scans.StartScan(r.Context(), appservice.StartScanRequest{
		Mode: request.Mode,
		CIDR: request.CIDR,
	})
	if err != nil {
		writeScanServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (server Server) handleListScans(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid limit %q", rawLimit))
			return
		}
		limit = parsedLimit
	}

	scans, err := server.repository.ListScans(r.Context(), storage.ListScansFilter{
		Query: r.URL.Query().Get("query"),
		Limit: limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, scans)
}

func (server Server) handleGetScan(w http.ResponseWriter, r *http.Request) {
	result, err := server.repository.GetScan(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, storage.ErrScanNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (server Server) handleExportScan(w http.ResponseWriter, r *http.Request) {
	result, err := server.repository.GetScan(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, storage.ErrScanNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	format := strings.ToLower(r.PathValue("format"))
	switch format {
	case "json":
		payload, err := report.JSON(result)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeDownload(w, http.StatusOK, "application/json", exportFilename(result, "json"), payload)
	case "csv":
		payload, err := report.CSV(result)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeDownload(w, http.StatusOK, "text/csv", exportFilename(result, "csv"), payload)
	default:
		writeError(w, http.StatusBadRequest, fmt.Errorf("unsupported export format %q", format))
	}
}

type startScanRequest struct {
	Mode string `json:"mode"`
	CIDR string `json:"cidr,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeScanServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, appservice.ErrUnsupportedScanMode):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, appservice.ErrLocalNetworkUnavailable):
		writeError(w, http.StatusServiceUnavailable, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeDownload(w http.ResponseWriter, status int, contentType string, filename string, payload []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func exportFilename(scan scanner.ScanResult, extension string) string {
	scanID := strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(scan.ID)
	if scanID == "" {
		scanID = "scan"
	}
	return fmt.Sprintf("netinspector-%s.%s", scanID, extension)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout, X-User-Agent")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, Grpc-Status, Grpc-Message, Connect-Protocol-Version")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
