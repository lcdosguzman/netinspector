# NetInspector

NetInspector is a local network discovery and inspection tool built with Go and a planned Next.js dashboard. The first MVP focuses on a controlled, portfolio-ready experience: automatic local network detection, a safe TCP-based discovery scan, and a demo mode that does not touch the real network.

## Current Status

Implemented:

- Go module using `github.com/lcdosguzman/netinspector`.
- CLI entrypoint at `cmd/netinspector`.
- Local IPv4 network detection.
- Demo scan with realistic sample devices.
- Controlled TCP discovery scan for common ports.
- Basic scanner tests.

Planned next:

- Connect-RPC API server.
- Protocol Buffers contract.
- Next.js dashboard.
- React Flow topology view.
- SQLite scan history.

## Run

Demo mode:

```sh
go run ./cmd/netinspector -mode demo
```

Real scan mode:

```sh
go run ./cmd/netinspector -mode real
```

The real scan detects the active local IPv4 network and probes common TCP ports. It avoids raw packets for the first MVP, which keeps the implementation more portable across macOS, Linux, and Windows.

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

The dashboard runs at `http://127.0.0.1:3000` by default.

## Test

```sh
go test ./...
```

Dashboard checks:

```sh
cd apps/dashboard
npm run lint
npm run build
```
