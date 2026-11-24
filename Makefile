GO ?= go
BIN_DIR := bin
SERVER_BIN := $(BIN_DIR)/llm-wiki-server

.PHONY: build migrate-up

# build compiles the HTTP server.
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(SERVER_BIN) ./cmd/server

# migrate-up runs the migration CLI directly.
migrate:
	$(GO) run ./cmd/migrate -direction up