# NetInspector Pending Improvements

This document lists only the improvements that are still pending for NetInspector.

## Clean Code

### 1. Remove Unused Files

Review `apps/dashboard/go.mod`. A Next.js dashboard normally should not need a Go module file inside its app folder unless there is a specific reason.

If unused, remove it to avoid confusion.

## Engineering Practices

### 1. Add More Tests

Useful test areas:

- mDNS hostname normalization and device type inference.
- SSDP response parsing edge cases.
- ARP parsing edge cases.
- local network detection helpers.
- Protobuf mapping functions.
- frontend data mappers.
- report export formatting.

### 2. Add Project Metadata

Useful repository files:

- `LICENSE`
- `CONTRIBUTING.md`
- `.env.example`
- short architecture diagram in README.

### 3. Improve Logging

Use Go `log/slog` for structured logs.

Useful events:

- scan started
- scan finished
- interface selected
- discovery source errors
- scan duration
- device count
- report exported

Structured logs make the backend feel more production-ready.

## Product Features

### 1. Device Change Detection

Compare the current scan against previous scans.

Useful events:

- new device detected
- device disappeared
- hostname changed
- MAC changed
- new port opened
- port closed

This turns the app from a one-time scanner into a useful monitoring tool.

### 2. Deep Device Inspection

Add an `Inspect` action per device.

Possible details:

- extended port scan
- service fingerprinting
- estimated OS
- additional hostname lookups
- risk or exposure hints

This can use the existing `InspectDevice` RPC already defined in the Protobuf contract.

### 3. Confidence Score

Show an identification confidence score.

Example:

- `TV - 87% confidence`
- Evidence: Google Cast service, vendor hint, port 8443, hostname.

This would make the device classification feel more transparent and professional.

### 4. Privacy-Aware Explanations

Add clearer UI explanations when:

- MAC addresses are randomized.
- OUI identifies the Wi-Fi chip vendor, not necessarily the product brand.
- mDNS/Bonjour data is unavailable.
- SSDP/UPnP data is unavailable.

This shows responsible engineering judgment.

### 5. Better Graph Layout

React Flow is already implemented, but the layout is deterministic and simple.

Future improvements:

- automatic layout with Dagre or ELK
- grouping by device type
- visual separation for router, clients, IoT, and printers
- color coding by confidence or status

### 6. PDF Report Export

JSON and CSV export already exist. A PDF summary would make the app more useful for sharing scan results with non-technical users.

Possible PDF sections:

- scan summary
- device table
- topology snapshot
- identification notes
- privacy limitations

## Recommended Next Order

1. Add device change detection.
2. Add deep device inspection.
3. Add confidence score.
4. Add privacy-aware explanations.
5. Improve graph layout.
6. Add PDF report export.
