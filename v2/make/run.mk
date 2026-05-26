# =========================================
# make/run.mk
# Run local API and web dev server
# =========================================

.PHONY: run-help run-all stop-all test-all build-all verify

ROOT_DIR ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
GC_API_DIR ?= $(ROOT_DIR)/gc-api
GC_WEB_DIR ?= $(ROOT_DIR)/gc-web
GC_API_CONFIG_FILE ?= $(GC_API_DIR)/config/config.yaml
GC_API_PORT ?= 2041
GC_API_ADDR ?= :$(GC_API_PORT)
GC_WEB_PORT ?= 2040
WEB_HOST ?= 0.0.0.0

run-help:
	@echo "Usage examples:"
	@echo "  make run-all"
	@echo "  make stop-all"
	@echo "  make test-all"
	@echo "  make build-all"
	@echo "  make verify"
	@echo "  make run-all GC_API_CONFIG_FILE=$(GC_API_CONFIG_FILE)"
	@echo "  make run-all WEB_HOST=$(WEB_HOST) GC_WEB_PORT=$(GC_WEB_PORT)"

run-all: stop-all
	@echo "Starting Rudy GC API and web dev server"
	@set -e; \
	(cd "$(GC_API_DIR)" && go run ./cmd/api -f "$(GC_API_CONFIG_FILE)" -addr "$(GC_API_ADDR)") & api_pid=$$!; \
	(cd "$(GC_WEB_DIR)" && npm run dev -- --host "$(WEB_HOST)" --port "$(GC_WEB_PORT)") & web_pid=$$!; \
	trap 'kill -9 $$api_pid $$web_pid 2>/dev/null || true' INT TERM EXIT; \
	wait $$api_pid $$web_pid

stop-all:
	@pids="$$(lsof -tiTCP:$(GC_API_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping Rudy GC API on port $(GC_API_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_API_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "Rudy GC API port $(GC_API_PORT) is not in use"; \
	fi
	@pids="$$(lsof -tiTCP:$(GC_WEB_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping Rudy GC web on port $(GC_WEB_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_WEB_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "Rudy GC web port $(GC_WEB_PORT) is not in use"; \
	fi

test-all:
	@cd "$(GC_API_DIR)" && go test ./...
	@cd "$(GC_WEB_DIR)" && npm run test -- --run

build-all:
	@cd "$(GC_API_DIR)" && go build ./...
	@cd "$(GC_WEB_DIR)" && npm run build

verify: test-all build-all
