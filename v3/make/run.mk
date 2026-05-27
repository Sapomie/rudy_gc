ifndef V3_RUN_MK
V3_RUN_MK := 1

V3_RUN_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_RUN_MAKE_DIR := $(dir $(V3_RUN_FILE))

include $(V3_RUN_MAKE_DIR)common.mk
include $(V3_RUN_MAKE_DIR)gc-api.mk
include $(V3_RUN_MAKE_DIR)gc-web.mk

.PHONY: run-help run-all stop-all

run-help:
	@echo "Run targets:"
	@echo "  make run-all"
	@echo "  make stop-all"
	@echo "Run variables:"
	@echo "  GC_API_DIR=$(GC_API_DIR)"
	@echo "  GC_WEB_DIR=$(GC_WEB_DIR)"
	@echo "  GC_API_PORT=$(GC_API_PORT)"
	@echo "  GC_WEB_PORT=$(GC_WEB_PORT)"

run-all: stop-all
	@echo "Starting v3 API and web"
	@set -e; \
	(cd "$(GC_API_DIR)" && go run ./cmd/api -f "$(GC_API_CONFIG_FILE)") & api_pid=$$!; \
	(cd "$(GC_WEB_DIR)" && PORT="$(GC_WEB_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev) & web_pid=$$!; \
	trap 'kill -9 $$api_pid $$web_pid 2>/dev/null || true' INT TERM EXIT; \
	wait $$api_pid $$web_pid

stop-all: stop-api stop-web

endif
