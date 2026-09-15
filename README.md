# NetInspector

[![CI](https://github.com/lcdosguzman/netinspector/actions/workflows/ci.yml/badge.svg)](https://github.com/lcdosguzman/netinspector/actions/workflows/ci.yml)

NetInspector is a local network discovery and inspection dashboard built with Go and Next.js. The MVP focuses on a controlled, portfolio-ready experience: automatic local network detection, demo data, real LAN discovery, device inspection, and a visual dashboard.

## What It Does

- Detects the active local IPv4 network.
- Runs a demo scan with realistic sample devices.
- Runs a real scan using controlled TCP probes, the local ARP cache, mDNS/Bonjour, and SSDP/UPnP service discovery.
- Shows discovered devices in a dashboard.
- Renders the topology with an interactive React Flow graph.
- Saves real scan history locally with SQLite.
- Exports saved scan reports as JSON or CSV.
- Displays IP, hostname, MAC, vendor, inferred device type, latency, open ports, and identification hints.
- Uses OUI-based MAC vendor lookup for known prefixes.
- Flags private/randomized MAC addresses, which are common on phones, tablets, and laptops with Wi-Fi privacy enabled.

## Architecture

```text
netinspector/
  cmd/netinspector/        Go CLI and API server entrypoint
  internal/api/            Connect-RPC service and compatibility JSON API
  internal/gen/            Generated Go Protobuf and Connect code
  internal/demo/           Demo scan data
  internal/network/        Local network detection
  internal/scanner/        Discovery interfaces, TCP probes, ARP parsing, mDNS and SSDP discovery, OUI/vendor enrichment
  internal/storage/        Scan history repository interface and SQLite implementation
  proto/network/v1/        Protobuf API contract
  apps/dashboard/          Next.js dashboard
  apps/dashboard/components/ Dashboard UI components
  apps/dashboard/lib/      Dashboard client and mapping helpers
  apps/dashboard/src/gen/  Generated TypeScript Protobuf and Connect client
```

## Current MVP

Implemented:

- Go module using `github.com/lcdosguzman/netinspector`.
- CLI scan modes: `demo` and `real`.
- Local IPv4 network detection.
- Demo scan with realistic sample devices.
- Real scan with TCP probes for common ports.
- ARP cache discovery to find devices that do not expose common TCP ports.
- mDNS/Bonjour discovery for `.local` hostnames and advertised services such as AirPlay, Google Cast, SMB, SSH, HTTP, and printers.
- SSDP/UPnP discovery for routers, smart TVs, media devices, consoles, printers, and IoT devices that advertise UPnP services.
- Scanner discovery sources behind a shared `Discoverer` interface.
- SQLite-backed scan history behind a `ScanRepository` interface.
- JSON and CSV report exports for saved scans.
- OUI/vendor enrichment for known MAC prefixes.
- Basic device type inference.
- Connect-RPC API generated from Protocol Buffers.
- Compatibility HTTP JSON endpoints.
- Interactive Next.js dashboard.
- React Flow topology graph with draggable nodes, zoom, pan, minimap, and selectable devices.
- TypeScript dashboard client generated from the same `.proto` contract used by Go.
- Makefile development shortcuts.
- GitHub Actions CI for generated code, backend tests, dashboard lint, and dashboard build.
- Scanner tests.

Known limitations:

- OUI identifies the network chip/vendor, not always the commercial product brand.
- Phones and laptops may use private/randomized MAC addresses, so their real manufacturer cannot always be inferred.
- mDNS/Bonjour only enriches devices that advertise services on the local network.
- SSDP/UPnP only enriches devices that respond to multicast UPnP discovery.
- The current real scanner is intentionally conservative and avoids raw packet capture for MVP portability.
- The topology layout is deterministic for the MVP; automated layout can be added as the graph grows.

## Run

Demo CLI:

```sh
go run ./cmd/netinspector -mode demo
```

Real CLI scan:

```sh
go run ./cmd/netinspector -mode real
```

API server:

```sh
go run ./cmd/netinspector -serve
```

Available endpoints:

- `GET /healthz`
- `POST /network.v1.NetworkService/StartScan` via Connect-RPC
- `POST /network.v1.NetworkService/ListScans` via Connect-RPC
- `POST /network.v1.NetworkService/GetScan` via Connect-RPC
- `POST /network.v1.NetworkService/GetLocalNetwork` via Connect-RPC
- `GET /api/local-network`
- `GET /api/scans`
- `GET /api/scans/{id}`
- `GET /api/scans/{id}/export/json`
- `GET /api/scans/{id}/export/csv`
- `POST /api/scans` with `{"mode":"DEMO"}` or `{"mode":"REAL"}`

Scan history is stored in SQLite at:

```text
.netinspector/netinspector.db
```

Override the database path with:

```sh
NETINSPECTOR_DB_PATH=/path/to/netinspector.db go run ./cmd/netinspector -serve
```

or:

```sh
go run ./cmd/netinspector -serve -db /path/to/netinspector.db
```

Generate Protobuf and Connect code:

```sh
npm install
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
PATH="$PATH:$(go env GOPATH)/bin:$(pwd)/node_modules/.bin" npm run proto:generate
```

Dashboard:

```sh
cd apps/dashboard
npm install
npm run dev
```

Open:

```text
http://127.0.0.1:3000
```

The dashboard expects the backend at:

```text
http://127.0.0.1:8088
```

## Test

Recommended local check:

```sh
make check
```

Backend:

```sh
go test ./...
```

Dashboard:

```sh
cd apps/dashboard
npm run lint
npm run build
```

Useful development shortcuts:

```sh
make install
make dev
make proto
make test
make lint
make build
```

## Roadmap

- Replace the seed OUI table with a complete local IEEE OUI database.
- Add deeper device inspection on demand.
- Improve cross-platform discovery for Linux and Windows.
