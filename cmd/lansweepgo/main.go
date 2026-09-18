package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lcdosguzman/lansweepgo/internal/api"
	"github.com/lcdosguzman/lansweepgo/internal/demo"
	"github.com/lcdosguzman/lansweepgo/internal/network"
	"github.com/lcdosguzman/lansweepgo/internal/scanner"
)

func main() {
	mode := flag.String("mode", "demo", "scan mode: demo or real")
	serve := flag.Bool("serve", false, "start the local API server")
	addr := flag.String("addr", "127.0.0.1:8088", "API server address")
	dbPath := flag.String("db", "", "SQLite database path for scan history")
	timeout := flag.Duration("timeout", 300*time.Millisecond, "timeout per host probe")
	flag.Parse()

	ctx := context.Background()
	if *serve {
		server := api.NewServer(api.Config{
			Addr:         *addr,
			ScanTimeout:  *timeout,
			DatabasePath: *dbPath,
		})
		log.Printf("lansweepgo API listening on http://%s", *addr)
		if err := server.ListenAndServe(ctx); err != nil {
			exitWithError(err)
		}
		return
	}

	switch *mode {
	case "demo":
		result := demo.NewScan()
		printJSON(result)
	case "real":
		local, err := network.DetectLocalNetwork()
		if err != nil {
			exitWithError(err)
		}

		scan := scanner.NewTCPScanner(scanner.Config{
			Ports:       scanner.DefaultDiscoveryPorts(),
			Concurrency: 128,
			Timeout:     *timeout,
		})

		result, err := scan.ScanCIDR(ctx, local)
		if err != nil {
			exitWithError(err)
		}

		printJSON(result)
	default:
		exitWithError(fmt.Errorf("unsupported mode %q: expected demo or real", *mode))
	}
}

func printJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "lansweepgo: %v\n", err)
	os.Exit(1)
}
