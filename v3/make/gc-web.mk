ifndef V3_GC_WEB_MK
V3_GC_WEB_MK := 1

V3_GC_WEB_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_GC_WEB_MAKE_DIR := $(dir $(V3_GC_WEB_FILE))

include $(V3_GC_WEB_MAKE_DIR)common.mk

.PHONY: web-help run-web stop-web check-web

web-help:
	@echo "Web targets:"
	@echo "  make run-web"
	@echo "  make stop-web"
	@echo "  make check-web"
	@echo "Web variables:"
	@echo "  GC_WEB_DIR=$(GC_WEB_DIR)"
	@echo "  GC_WEB_PORT=$(GC_WEB_PORT)"
	@echo "  GC_API_ORIGIN=$(GC_API_ORIGIN)"

run-web: stop-web
	@echo "Starting v3 web from $(GC_WEB_DIR)"
	@cd "$(GC_WEB_DIR)" && PORT="$(GC_WEB_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev

stop-web:
	@pids="$$(lsof -tiTCP:$(GC_WEB_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping v3 web on port $(GC_WEB_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_WEB_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "v3 web port $(GC_WEB_PORT) is not in use"; \
	fi

check-web:
	@cd "$(GC_WEB_DIR)" && npm run check

endif
