# Contributing

Thanks for your interest in improving NetInspector.

## Development Setup

Install dependencies:

```sh
make install
```

Run the API and dashboard locally:

```sh
make dev
```

The API runs on `127.0.0.1:8088` and the dashboard runs on `127.0.0.1:3000` by default.

## Before Opening A PR

Run the local checks:

```sh
make test
make lint
make build
```

If you change the Protobuf contract, regenerate generated code:

```sh
make proto
```

## Code Guidelines

- Keep backend business logic out of transport handlers when possible.
- Prefer small packages and files with clear responsibilities.
- Use table-driven tests for Go test cases.
- Keep frontend components focused and typed.
- Avoid committing local scan databases, `.env` files, or planning notes.

## Security And Privacy

NetInspector inspects local networks. Do not commit real scan outputs that expose private IPs, MAC addresses, hostnames, or device names.
