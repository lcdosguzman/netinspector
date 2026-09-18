SHELL := /bin/bash

API_ADDR ?= 127.0.0.1:8088
DASHBOARD_HOST ?= 127.0.0.1
DASHBOARD_PORT ?= 3000
GO_BIN := $(shell go env GOPATH)/bin
PROTO_PATH := $(PATH):$(GO_BIN):$(CURDIR)/node_modules/.bin
GOCACHE ?= $(CURDIR)/.cache/go-build

.PHONY: help install test test-go test-dashboard lint build proto proto-check check dev dev-api dev-dashboard

help:
	@echo "LanSweepGo development commands"
	@echo ""
	@echo "  make install       Install root and dashboard npm dependencies"
	@echo "  make test          Run Go and dashboard tests"
	@echo "  make lint          Run dashboard lint"
	@echo "  make build         Build dashboard"
	@echo "  make proto         Regenerate Protobuf and Connect code"
	@echo "  make proto-check   Regenerate Protobuf code and fail if generated files changed"
	@echo "  make check         Run proto-check, tests, lint, and build"
	@echo "  make dev           Run API and dashboard locally"
	@echo "  make dev-api       Run only the API server"
	@echo "  make dev-dashboard Run only the dashboard"

install:
	npm install
	cd apps/dashboard && npm install

test:
	$(MAKE) test-go
	$(MAKE) test-dashboard

test-go:
	GOCACHE="$(GOCACHE)" go test ./cmd/... ./internal/...

test-dashboard:
	cd apps/dashboard && npm run test

lint:
	cd apps/dashboard && npm run lint

build:
	cd apps/dashboard && npm run build

proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	PATH="$(PROTO_PATH)" npm run proto:generate

proto-check: proto
	git diff --exit-code -- proto internal/gen apps/dashboard/src/gen

check: proto-check test lint build

dev-api:
	go run ./cmd/lansweepgo -serve -addr $(API_ADDR)

dev-dashboard:
	cd apps/dashboard && npm run dev -- --hostname $(DASHBOARD_HOST) --port $(DASHBOARD_PORT)

dev:
	@echo "Starting API on http://$(API_ADDR) and dashboard on http://$(DASHBOARD_HOST):$(DASHBOARD_PORT)"
	@trap 'kill $$(jobs -p)' EXIT INT TERM; \
	go run ./cmd/lansweepgo -serve -addr $(API_ADDR) & \
	cd apps/dashboard && npm run dev -- --hostname $(DASHBOARD_HOST) --port $(DASHBOARD_PORT)
