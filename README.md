# LanSweepGo

[![CI](https://github.com/lcdosguzman/lansweepgo/actions/workflows/ci.yml/badge.svg)](https://github.com/lcdosguzman/lansweepgo/actions/workflows/ci.yml)

LanSweepGo is a local network discovery and inspection dashboard built with Go and Next.js. The MVP focuses on a controlled, portfolio-ready experience: automatic local network detection, demo data, real LAN discovery, device inspection, scan history, report exports, and an interactive visual dashboard.

## Tech Stack

- Go backend with a small CLI and local API server.
- Connect-RPC and Protocol Buffers for typed Go/TypeScript API contracts.
- Next.js, React, TypeScript, and React Flow for the dashboard.
- SQLite for local scan history.
- Vitest, Go tests, ESLint, Buf, Makefile shortcuts, and GitHub Actions CI.

## What It Does

- Detects the active local IPv4 network.
- Runs a demo scan with realistic sample devices.
- Runs a real scan using controlled TCP probes, the local ARP cache, mDNS/Bonjour, and SSDP/UPnP service discovery.
- Shows discovered devices in a dashboard.
- Renders the topology with an interactive React Flow graph.
- Saves real scan history locally with SQLite.
- Lets users search and reload previous scans from the dashboard.
- Exports saved scan reports as JSON or CSV.
- Displays IP, hostname, MAC, vendor, inferred device type, latency, open ports, and identification hints.
- Uses OUI-based MAC vendor lookup for known prefixes.
- Flags private/randomized MAC addresses, which are common on phones, tablets, and laptops with Wi-Fi privacy enabled.

## Architecture

LanSweepGo is organized as a contract-first Go and Next.js application. The Go backend handles network discovery, device enrichment, scan orchestration, persistence, exports, and the local API. The Next.js dashboard consumes that API through TypeScript clients generated from the same Protocol Buffers contract used by the Go server.

The scanner is split into discovery sources behind interfaces. TCP probes, ARP cache parsing, mDNS/Bonjour, and SSDP/UPnP each provide device evidence that the scan service combines into one normalized result. Scan history is stored through a repository interface, with SQLite as the current implementation.

The dashboard covers the user workflows: starting scans, switching between demo and real mode, displaying the topology, inspecting devices, loading scan history, and exporting reports.

```text
lansweepgo/
  cmd/lansweepgo/        Go CLI and API server entrypoint
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

- Go module using `github.com/lcdosguzman/lansweepgo`.
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

Requirements:

- Go 1.26+
- Node.js 20+
- npm

Install dependencies:

```sh
make install
```

Run the API and dashboard together:

```sh
make dev
```

Open:

```text
http://127.0.0.1:3000
```

The API runs at:

```text
http://127.0.0.1:8088
```

Optional environment setup:

```sh
cp .env.example .env
```

Demo CLI:

```sh
go run ./cmd/lansweepgo -mode demo
```

Real CLI scan:

```sh
go run ./cmd/lansweepgo -mode real
```

API server:

```sh
go run ./cmd/lansweepgo -serve
```

Run only one side during development:

```sh
make dev-api
make dev-dashboard
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

By default, scan history is stored in SQLite at:

```text
.lansweepgo/lansweepgo.db
```

Override the database path with:

```sh
LANSWEEPGO_DB_PATH=/path/to/lansweepgo.db go run ./cmd/lansweepgo -serve
```

or:

```sh
go run ./cmd/lansweepgo -serve -db /path/to/lansweepgo.db
```

Generate Protobuf and Connect code:

```sh
npm install
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
PATH="$PATH:$(go env GOPATH)/bin:$(pwd)/node_modules/.bin" npm run proto:generate
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

Backend and dashboard tests:

```sh
make test
```

Dashboard:

```sh
cd apps/dashboard
npm run test
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
make check
```

## Notes For Portfolio Review

- The API is contract-first: `proto/network/v1/network.proto` generates both Go server types and TypeScript dashboard clients.
- Discovery is split behind interfaces so new sources can be added without rewriting the scanner orchestration.
- Scan persistence uses a repository interface, so SQLite can be replaced later by another storage engine with less impact.
- The real scanner is intentionally conservative for an MVP and avoids privileged packet capture.

## Roadmap

- Replace the seed OUI table with a complete local IEEE OUI database.
- Add deeper device inspection on demand.
- Improve cross-platform discovery for Linux and Windows.

## License

MIT. See [LICENSE](LICENSE).
