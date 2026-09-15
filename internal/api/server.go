package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lcdosguzman/netinspector/internal/demo"
	networkv1connect "github.com/lcdosguzman/netinspector/internal/gen/network/v1/networkv1connect"
	"github.com/lcdosguzman/netinspector/internal/network"
	"github.com/lcdosguzman/netinspector/internal/scanner"
)

type Config struct {
	Addr        string
	ScanTimeout time.Duration
}

type Server struct {
	config Config
}

func NewServer(config Config) Server {
	if config.Addr == "" {
		config.Addr = "127.0.0.1:8088"
	}
	if config.ScanTimeout <= 0 {
		config.ScanTimeout = 300 * time.Millisecond
	}

	return Server{config: config}
}

func (server Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /api/local-network", server.handleLocalNetwork)
	mux.HandleFunc("POST /api/scans", server.handleStartScan)
	path, handler := networkv1connect.NewNetworkServiceHandler(newNetworkService(server.config))
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

	switch request.Mode {
	case "", "DEMO":
		writeJSON(w, http.StatusOK, demo.NewScan())
	case "REAL":
		local, err := network.DetectLocalNetwork()
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err)
			return
		}
		if request.CIDR != "" {
			local.CIDR = request.CIDR
		}

		tcpScanner := scanner.NewTCPScanner(scanner.Config{
			Ports:       scanner.DefaultDiscoveryPorts(),
			Concurrency: 128,
			Timeout:     server.config.ScanTimeout,
		})
		result, err := tcpScanner.ScanCIDR(r.Context(), local)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	default:
		writeError(w, http.StatusBadRequest, fmt.Errorf("unsupported scan mode %q", request.Mode))
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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout, X-User-Agent")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message, Connect-Protocol-Version")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
