ifndef V3_GC_APPS_MK
V3_GC_APPS_MK := 1

V3_GC_APPS_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_GC_APPS_MAKE_DIR := $(dir $(V3_GC_APPS_FILE))

include $(V3_GC_APPS_MAKE_DIR)common.mk

GC_API_CONFIG_FILE ?= $(GC_API_DIR)/cmd/api/config.yaml
GC_WEB_NEXT02_DIR ?= $(ROOT_DIR)/gc-web-next02
GC_WEB_NEXT02_PORT ?= 5022

.PHONY: apps-help api-help web-help web-next-help \
	run-api stop-api test-api build-api migrate-api \
	run-web-next02 stop-web-next02 check-web-next02 \
	run-web stop-web check-web \
	run-next stop-next check-next \
	run-dev stop-dev check-dev

apps-help:
	@echo "App targets:"
	@echo "  make run-api          Run v3 API on $(GC_API_PORT)"
	@echo "  make stop-api         Stop v3 API port"
	@echo "  make test-api         Run Go tests"
	@echo "  make build-api        Build Go packages"
	@echo "  make migrate-api      Run database migrations"
	@echo "  make run-web          Run gc-web-next02 on $(GC_WEB_NEXT02_PORT)"
	@echo "  make stop-web         Stop gc-web-next02 port"
	@echo "  make check-web        Check gc-web-next02"
	@echo "  make run-web-next02   Run gc-web-next02 on $(GC_WEB_NEXT02_PORT)"
	@echo "  make stop-web-next02  Stop gc-web-next02 port"
	@echo "  make check-web-next02 Check gc-web-next02"
	@echo "  make run-next         Alias of run-web-next02"
	@echo "  make stop-next        Alias of stop-web-next02"
	@echo "  make check-next       Alias of check-web-next02"
	@echo "  make run-dev          Run API and gc-web-next02 together"
	@echo "  make stop-dev         Stop API and gc-web-next02"
	@echo "  make check-dev        Test API and check gc-web-next02"
	@echo "App variables:"
	@echo "  GC_API_DIR=$(GC_API_DIR)"
	@echo "  GC_API_CONFIG_FILE=$(GC_API_CONFIG_FILE)"
	@echo "  GC_API_PORT=$(GC_API_PORT)"
	@echo "  GC_WEB_NEXT02_DIR=$(GC_WEB_NEXT02_DIR)"
	@echo "  GC_WEB_NEXT02_PORT=$(GC_WEB_NEXT02_PORT)"
	@echo "  GC_API_ORIGIN=$(GC_API_ORIGIN)"

api-help: apps-help

web-help: apps-help

web-next-help: apps-help

run-api: stop-api
	@echo "Starting v3 API from $(GC_API_DIR)"
	@cd "$(GC_API_DIR)" && go run ./cmd/api -f "$(GC_API_CONFIG_FILE)"

stop-api:
	@pids="$$(lsof -tiTCP:$(GC_API_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping v3 API on port $(GC_API_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_API_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "v3 API port $(GC_API_PORT) is not in use"; \
	fi

test-api:
	@cd "$(GC_API_DIR)" && go test ./...

build-api:
	@cd "$(GC_API_DIR)" && go build ./...

migrate-api:
	@cd "$(GC_API_DIR)" && go run ./cmd/migrate -f "$(GC_API_CONFIG_FILE)"

run-web-next02: stop-web-next02
	@echo "Starting gc-web-next02 from $(GC_WEB_NEXT02_DIR)"
	@cd "$(GC_WEB_NEXT02_DIR)" && PORT="$(GC_WEB_NEXT02_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev

stop-web-next02:
	@pids="$$(lsof -tiTCP:$(GC_WEB_NEXT02_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping gc-web-next02 on port $(GC_WEB_NEXT02_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_WEB_NEXT02_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "gc-web-next02 port $(GC_WEB_NEXT02_PORT) is not in use"; \
	fi

check-web-next02:
	@cd "$(GC_WEB_NEXT02_DIR)" && npm run check

run-web: run-web-next02

stop-web: stop-web-next02

check-web: check-web-next02

run-next: run-web-next02

stop-next: stop-web-next02

check-next: check-web-next02

run-dev: stop-dev
	@echo "Starting v3 API and gc-web-next02"
	@set -e; \
	(cd "$(GC_API_DIR)" && go run ./cmd/api -f "$(GC_API_CONFIG_FILE)") & api_pid=$$!; \
	(cd "$(GC_WEB_NEXT02_DIR)" && PORT="$(GC_WEB_NEXT02_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev) & next_pid=$$!; \
	trap 'kill -9 $$api_pid $$next_pid 2>/dev/null || true' INT TERM EXIT; \
	wait $$api_pid $$next_pid

stop-dev: stop-api stop-web-next02

check-dev: test-api check-web-next02

endif
