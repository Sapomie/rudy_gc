ifndef REMOTE_PROJECT_APPS_MK
REMOTE_PROJECT_APPS_MK := 1

REMOTE_APPS_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
REMOTE_APPS_MAKE_DIR := $(dir $(REMOTE_APPS_FILE))

include $(REMOTE_APPS_MAKE_DIR)common.mk

.PHONY: run-api stop-api test-api build-api migrate-api \
	run-web stop-web check-web test-web build-web \
	run-dev stop-dev stop-all

run-api: stop-api
	@cd "$(API_DIR)" && go run ./cmd/api -f "$(API_CONFIG_FILE)"

stop-api:
	@pids="$$(lsof -tiTCP:$(API_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping remote API on port $(API_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(API_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "Remote API port $(API_PORT) is not in use"; \
	fi

test-api:
	@cd "$(API_DIR)" && go test ./...

build-api:
	@cd "$(API_DIR)" && go build ./...

migrate-api:
	@cd "$(API_DIR)" && go run ./cmd/migrate -f "$(API_CONFIG_FILE)"

run-web: stop-web
	@cd "$(WEB_DIR)" && PORT="$(WEB_PORT)" API_ORIGIN="$(API_ORIGIN)" npm run dev

stop-web:
	@pids="$$(lsof -tiTCP:$(WEB_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping remote web on port $(WEB_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(WEB_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "Remote web port $(WEB_PORT) is not in use"; \
	fi

check-web:
	@cd "$(WEB_DIR)" && npm run check

test-web:
	@cd "$(WEB_DIR)" && npm test

build-web:
	@cd "$(WEB_DIR)" && npm run build

run-dev: stop-dev
	@set -e; \
	(cd "$(API_DIR)" && go run ./cmd/api -f "$(API_CONFIG_FILE)") & api_pid=$$!; \
	(cd "$(WEB_DIR)" && PORT="$(WEB_PORT)" API_ORIGIN="$(API_ORIGIN)" npm run dev) & web_pid=$$!; \
	trap 'kill -9 $$api_pid $$web_pid 2>/dev/null || true' INT TERM EXIT; \
	wait $$api_pid $$web_pid

stop-dev: stop-api stop-web

stop-all: stop-api stop-web

endif
