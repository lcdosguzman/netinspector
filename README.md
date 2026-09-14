# NetInspector

NetInspector is a local network discovery and inspection dashboard built with Go and Next.js. The MVP focuses on a controlled, portfolio-ready experience: automatic local network detection, demo data, real LAN discovery, device inspection, and a visual dashboard.

## What It Does

- Detects the active local IPv4 network.
- Runs a demo scan with realistic sample devices.
- Runs a real scan using controlled TCP probes and the local ARP cache.
- Shows discovered devices in a dashboard.
- Displays IP, MAC, vendor, inferred device type, latency, open ports, and identification hints.
- Uses OUI-based MAC vendor lookup for known prefixes.
- Flags private/randomized MAC addresses, which are common on phones, tablets, and laptops with Wi-Fi privacy enabled.

## Architecture

```text
netinspector/
  cmd/netinspector/        Go CLI and API server entrypoint
  internal/api/            HTTP API for local scans
  internal/demo/           Demo scan data
  internal/network/        Local network detection
  internal/scanner/        TCP probes, ARP parsing, OUI/vendor enrichment
  proto/network/v1/        Future typed API contract
  apps/dashboard/          Next.js dashboard
```

## Current MVP

Implemented:

- Go module using `github.com/lcdosguzman/netinspector`.
- CLI scan modes: `demo` and `real`.
- Local IPv4 network detection.
- Demo scan with realistic sample devices.
- Real scan with TCP probes for common ports.
- ARP cache discovery to find devices that do not expose common TCP ports.
- OUI/vendor enrichment for known MAC prefixes.
- Basic device type inference.
- HTTP API server.
- Interactive Next.js dashboard.
- Scanner tests.

Known limitations:

- OUI identifies the network chip/vendor, not always the commercial product brand.
- Phones and laptops may use private/randomized MAC addresses, so their real manufacturer cannot always be inferred.
- The current real scanner is intentionally conservative and avoids raw packet capture for MVP portability.
- The dashboard graph is custom/static positioning for now; React Flow integration is planned.

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
- `GET /api/local-network`
- `POST /api/scans` with `{"mode":"DEMO"}` or `{"mode":"REAL"}`

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

## Roadmap

- Replace the seed OUI table with a complete local IEEE OUI database.
- Add mDNS/Bonjour hostname discovery.
- Add React Flow for a richer topology graph.
- Add SQLite scan history.
- Add Connect-RPC / Protocol Buffers API generation.
- Add deeper device inspection on demand.
- Improve cross-platform discovery for Linux and Windows.

