# NetInspector Project Improvements

This document captures technical and product improvements that would make NetInspector cleaner, easier to maintain, and more attractive as a portfolio project.

Status legend:

- `Done`: already implemented.
- `Pending`: still recommended.

## Current Strengths

- Clear Go backend structure using `cmd/`, `internal/`, and domain-oriented packages.
- Modern API contract with Protocol Buffers and Connect-RPC.
- Next.js dashboard with generated TypeScript client code.
- Interactive topology graph using React Flow.
- Real network discovery through TCP probing, ARP cache parsing, mDNS/Bonjour enrichment, and SSDP/UPnP discovery.
- Saved scan export in JSON and CSV formats.
- Basic scanner test coverage.
- Functional README and `.gitignore` that excludes local planning documents.

## Clean Code Improvements

### 1. Split Scanner Responsibilities

`internal/scanner/tcp.go` currently contains scan orchestration, TCP probing, device merging, IP helpers, and service-name helpers.

Suggested structure:

- `scanner/scanner.go`: scan orchestration.
- `scanner/tcp_probe.go`: TCP port probing.
- `scanner/discovery.go`: shared discovery interfaces.
- `scanner/ip.go`: IP/CIDR helper functions.
- `scanner/ports.go`: default ports and service names.

This would make the scanner easier to extend and test.

### 2. Add Discovery Interfaces

Status: `Done`

A common interface now defines scanner discovery sources:

```go
type Discoverer interface {
	Discover(ctx context.Context, local network.LocalNetwork) ([]Device, error)
}
```

Current implementations:

- `TCPDiscoverer`
- `ARPDiscoverer`
- `MDNSDiscoverer`
- `SSDPDiscoverer`

This would make the architecture more extensible and better aligned with clean architecture principles.

### 3. Centralize Scan Use Case Logic

The JSON compatibility API and the Connect-RPC service currently contain similar scan orchestration logic.

Recommended improvement:

- Create an internal application service, for example `internal/service/scan_service.go`.
- Let both Connect-RPC and JSON handlers call that service.
- Keep transport-specific code in the API layer only.

This reduces duplication and makes the backend easier to evolve.

### 4. Split The Dashboard Page

Status: `Done`

`apps/dashboard/app/page.tsx` was doing too many things: state management, API calls, topology rendering, device inspector, event console, and data mapping.

Implemented frontend structure:

- `components/ScanControls.tsx`
- `components/NetworkSummary.tsx`
- `components/TopologyGraph.tsx`
- `components/DeviceInspector.tsx`
- `components/EventConsole.tsx`
- `lib/netinspector-client.ts`
- `lib/mappers.ts`
- `types/network.ts`

This would make the frontend more maintainable and easier to test.

### 5. Remove Unused Files

Review `apps/dashboard/go.mod`. A Next.js dashboard normally should not need a Go module file inside its app folder unless there is a specific reason.

If unused, remove it to avoid confusion.

## Engineering Practice Improvements

### 1. Add A Makefile

Status: `Done`

A `Makefile` now provides simple standard commands for contributors and for local development.

Available commands:

```sh
make dev
make test
make lint
make build
make proto
make check
```

This improves developer experience and makes the repository look more professional.

### 2. Add GitHub Actions CI

Status: `Done`

CI now runs on pull requests and pushes to `main`:

- `go test ./...`
- dashboard lint
- dashboard build
- Protobuf generation validation

This shows professional delivery discipline and protects the project from regressions.

### 3. Add More Tests

Useful test areas:

- mDNS hostname normalization and device type inference.
- ARP parsing edge cases.
- local network detection helpers.
- Protobuf mapping functions.
- frontend data mappers.

### 4. Add Project Metadata

Useful repository files:

- `LICENSE`
- `CONTRIBUTING.md`
- `.env.example`
- GitHub Actions badge in README.
- Short architecture diagram in README.

### 5. Improve Logging

Use Go `log/slog` for structured logs.

Examples:

- scan started
- scan finished
- interface selected
- discovery source errors
- scan duration
- device count

Structured logs make the backend feel more production-ready.

## Product Improvements

### 1. SQLite Scan History

Status: `Done`

Previous real scans are now stored locally with SQLite.

This enables:

- scan history
- comparison between scans
- device first seen / last seen
- persistent dashboard data

This is one of the strongest next improvements for portfolio value.

### 2. Device Change Detection

Compare current scans against previous scans.

Useful events:

- new device detected
- device disappeared
- hostname changed
- MAC changed
- new port opened
- port closed

This turns the app from a one-time scanner into a useful monitoring tool.

### 3. Deep Device Inspection

Add an `Inspect` action per device.

Possible details:

- extended port scan
- service fingerprinting
- estimated OS
- additional hostname lookups
- risk or exposure hints

This can use the existing `InspectDevice` RPC already defined in the Protobuf contract.

### 4. Confidence Score

Show an identification confidence score.

Example:

- `TV - 87% confidence`
- Evidence: Google Cast service, vendor hint, port 8443, hostname.

This would make the device classification feel more transparent and professional.

### 5. SSDP/UPnP Discovery

Status: `Done`

SSDP discovery now complements mDNS/Bonjour by sending a controlled UPnP multicast search and merging responses into the scan result.

This helps identify:

- smart TVs
- routers
- consoles
- media devices
- IoT devices

### 6. Report Export

Status: `Done`

Saved scans can now be exported from the backend and downloaded from the dashboard.

Implemented formats:

- JSON
- CSV

Future format:

- PDF summary

This makes the app more useful for real-world diagnostics.

### 7. Privacy-Aware Explanations

Add clearer UI explanations when:

- MAC addresses are randomized.
- OUI identifies the Wi-Fi chip vendor, not necessarily the product brand.
- mDNS/Bonjour data is unavailable.

This shows responsible engineering judgment.

### 8. Better Graph Layout

React Flow is already implemented, but the layout is deterministic and simple.

Future improvements:

- automatic layout with Dagre or ELK
- grouping by device type
- visual separation for router, clients, IoT, and printers
- color coding by confidence or status

## Recommended Next Steps

Completed:

1. Split the dashboard into components.
2. Add a `Makefile`.
3. Add GitHub Actions CI.
4. Refactor scanner discovery sources behind interfaces.
5. Add SQLite scan history.
6. Add SSDP/UPnP discovery.
7. Add JSON and CSV report export.

Suggested next order:

1. Add device change detection.
2. Add deep device inspection.
3. Add PDF report export.

This order improves maintainability first, then adds professional product functionality.
