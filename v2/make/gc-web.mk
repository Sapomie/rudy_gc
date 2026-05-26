# =========================================
# make/gc-web.mk
# Rudy GC web local helpers
# =========================================

.PHONY: web-help install-web run-web stop-web test-web build-web preview-web

ROOT_DIR ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
GC_WEB_DIR ?= $(ROOT_DIR)/gc-web
GC_WEB_PORT ?= 2040
WEB_HOST ?= 0.0.0.0

web-help:
	@echo "Usage examples:"
	@echo "  make install-web"
	@echo "  make run-web"
	@echo "  make stop-web"
	@echo "  make test-web"
	@echo "  make build-web"
	@echo "  make preview-web"
	@echo "  make run-web WEB_HOST=$(WEB_HOST) GC_WEB_PORT=$(GC_WEB_PORT)"

install-web:
	@npm --prefix "$(GC_WEB_DIR)" install

run-web: stop-web
	@echo "Starting Rudy GC web dev server in: $(GC_WEB_DIR)"
	@cd "$(GC_WEB_DIR)" && npm run dev -- --host "$(WEB_HOST)" --port "$(GC_WEB_PORT)"

stop-web:
	@pids="$$(lsof -tiTCP:$(GC_WEB_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping Rudy GC web on port $(GC_WEB_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_WEB_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "Rudy GC web port $(GC_WEB_PORT) is not in use"; \
	fi

test-web:
	@cd "$(GC_WEB_DIR)" && npm run test -- --run

build-web:
	@cd "$(GC_WEB_DIR)" && npm run build

preview-web:
	@cd "$(GC_WEB_DIR)" && npm run preview -- --host "$(WEB_HOST)" --port "$(GC_WEB_PORT)"
