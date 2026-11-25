GO ?= go

.PHONY: build-api buid-mcp migrate-up

# build compiles the HTTP server.
build-api:
	@mkdir -p bin
	go build -o bin/api-server ./cmd/server

build-mcp:
	@mkdir -p bin
	go build -o bin/mcp-server ./cmd/mcp

# migrate-up runs the migration CLI directly.
migrate-up:
	$(GO) run ./cmd/migrate -direction up