ifndef V3_GC_API_MK
V3_GC_API_MK := 1

V3_GC_API_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_GC_API_MAKE_DIR := $(dir $(V3_GC_API_FILE))

include $(V3_GC_API_MAKE_DIR)common.mk

.PHONY: api-help run-api stop-api test-api build-api

GC_API_CONFIG_FILE ?= $(GC_API_DIR)/cmd/api/config.yaml

api-help:
	@echo "API targets:"
	@echo "  make run-api"
	@echo "  make stop-api"
	@echo "  make test-api"
	@echo "  make build-api"
	@echo "API variables:"
	@echo "  GC_API_DIR=$(GC_API_DIR)"
	@echo "  GC_API_CONFIG_FILE=$(GC_API_CONFIG_FILE)"
	@echo "  GC_API_PORT=$(GC_API_PORT)"

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

endif
