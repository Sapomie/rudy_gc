ifndef PROJECT_APPS_MK
PROJECT_APPS_MK := 1

PROJECT_APPS_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_APPS_MAKE_DIR := $(dir $(PROJECT_APPS_FILE))

include $(PROJECT_APPS_MAKE_DIR)common.mk

.PHONY: apps-help api-help web-help \
	run-api stop-api test-api build-api migrate-api \
	run-web stop-web check-web test-web build-web \
	run-dev stop-dev check-dev run-all stop-all

apps-help:
	@echo "App targets:"
	@echo "  make run-api          Run API on $(API_PORT)"
	@echo "  make stop-api         Stop API port"
	@echo "  make test-api         Run Go tests"
	@echo "  make build-api        Build Go packages"
	@echo "  make migrate-api      Run database migrations"
	@echo "  make run-web          Run web on $(WEB_PORT)"
	@echo "  make stop-web         Stop web port"
	@echo "  make check-web        Type-check web"
	@echo "  make test-web         Run web unit tests"
	@echo "  make build-web        Build web"
	@echo "  make run-dev          Run API and web together"
	@echo "  make stop-dev         Stop API and web"
	@echo "  make check-dev        Test API and check web"
	@echo "App variables:"
	@echo "  API_DIR=$(API_DIR)"
	@echo "  API_CONFIG_FILE=$(API_CONFIG_FILE)"
	@echo "  API_PORT=$(API_PORT)"
	@echo "  WEB_DIR=$(WEB_DIR)"
	@echo "  WEB_PORT=$(WEB_PORT)"
	@echo "  API_ORIGIN=$(API_ORIGIN)"

api-help: apps-help

web-help: apps-help

run-api: stop-api
	@echo "Starting API from $(API_DIR)"
	@cd "$(API_DIR)" && go run ./cmd/api -f "$(API_CONFIG_FILE)"

stop-api:
	@pids="$$(lsof -tiTCP:$(API_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping API on port $(API_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(API_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "API port $(API_PORT) is not in use"; \
	fi

test-api:
	@cd "$(API_DIR)" && go test ./...

build-api:
	@cd "$(API_DIR)" && go build ./...

migrate-api:
	@cd "$(API_DIR)" && go run ./cmd/migrate -f "$(API_CONFIG_FILE)"

run-web: stop-web
	@echo "Starting web from $(WEB_DIR)"
	@cd "$(WEB_DIR)" && PORT="$(WEB_PORT)" API_ORIGIN="$(API_ORIGIN)" npm run dev

stop-web:
	@pids="$$(lsof -tiTCP:$(WEB_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping web on port $(WEB_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(WEB_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "web port $(WEB_PORT) is not in use"; \
	fi


check-web:
	@cd "$(WEB_DIR)" && npm run check

test-web:
	@cd "$(WEB_DIR)" && npm test

build-web:
	@cd "$(WEB_DIR)" && npm run build

run-dev: stop-dev
	@echo "Starting API and web"
	@set -e; \
	(cd "$(API_DIR)" && go run ./cmd/api -f "$(API_CONFIG_FILE)") & api_pid=$$!; \
	(cd "$(WEB_DIR)" && PORT="$(WEB_PORT)" API_ORIGIN="$(API_ORIGIN)" npm run dev) & web_pid=$$!; \
	trap 'kill -9 $$api_pid $$web_pid 2>/dev/null || true' INT TERM EXIT; \
	wait $$api_pid $$web_pid

stop-dev: stop-api stop-web

check-dev: test-api check-web

run-all: run-dev

stop-all: stop-dev

endif
