ifndef V3_GC_WEB_NEXT_MK
V3_GC_WEB_NEXT_MK := 1

V3_GC_WEB_NEXT_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_GC_WEB_NEXT_MAKE_DIR := $(dir $(V3_GC_WEB_NEXT_FILE))

include $(V3_GC_WEB_NEXT_MAKE_DIR)common.mk

GC_WEB_NEXT01_DIR ?= $(ROOT_DIR)/gc-web-next01
GC_WEB_NEXT02_DIR ?= $(ROOT_DIR)/gc-web-next02
GC_WEB_NEXT03_DIR ?= $(ROOT_DIR)/gc-web-next03
GC_WEB_NEXT04_DIR ?= $(ROOT_DIR)/gc-web-next04
GC_WEB_NEXT05_DIR ?= $(ROOT_DIR)/gc-web-next05

GC_WEB_NEXT01_PORT ?= 5021
GC_WEB_NEXT02_PORT ?= 5022
GC_WEB_NEXT03_PORT ?= 5023
GC_WEB_NEXT04_PORT ?= 5024
GC_WEB_NEXT05_PORT ?= 5025

.PHONY: web-next-help \
	run-web-next01 run-web-next02 run-web-next03 run-web-next04 run-web-next05 run-web-next-all \
	stop-web-next01 stop-web-next02 stop-web-next03 stop-web-next04 stop-web-next05 stop-web-next-all \
	check-web-next02 check-web-next03 check-web-next04 check-web-next05 check-web-next-all \
	run-next stop-next check-next

web-next-help:
	@echo "Web next targets:"
	@echo "  make run-web-next02     Run gc-web-next02 on $(GC_WEB_NEXT02_PORT)"
	@echo "  make run-web-next03     Run gc-web-next03 on $(GC_WEB_NEXT03_PORT)"
	@echo "  make run-web-next04     Run gc-web-next04 on $(GC_WEB_NEXT04_PORT)"
	@echo "  make run-web-next05     Run gc-web-next05 on $(GC_WEB_NEXT05_PORT)"
	@echo "  make run-web-next-all   Run gc-web-next02..05 Vue frontends"
	@echo "  make stop-web-next-all  Stop gc-web-next02..05 frontend ports"
	@echo "  make check-web-next-all Check gc-web-next02..05"
	@echo "  make run-next           Alias of run-web-next-all"
	@echo "  make stop-next          Alias of stop-web-next-all"
	@echo "  make check-next         Alias of check-web-next-all"
	@echo "Web next variables:"
	@echo "  GC_WEB_NEXT01_DIR=$(GC_WEB_NEXT01_DIR)"
	@echo "  GC_WEB_NEXT02_DIR=$(GC_WEB_NEXT02_DIR)"
	@echo "  GC_WEB_NEXT03_DIR=$(GC_WEB_NEXT03_DIR)"
	@echo "  GC_WEB_NEXT04_DIR=$(GC_WEB_NEXT04_DIR)"
	@echo "  GC_WEB_NEXT05_DIR=$(GC_WEB_NEXT05_DIR)"
	@echo "  GC_API_ORIGIN=$(GC_API_ORIGIN)"

define RUN_WEB_NEXT_TARGET
run-web-next$(1): stop-web-next$(1)
	@echo "Starting gc-web-next$(1) from $$(GC_WEB_NEXT$(1)_DIR)"
	@cd "$$(GC_WEB_NEXT$(1)_DIR)" && PORT="$$(GC_WEB_NEXT$(1)_PORT)" V3_API_ORIGIN="$$(GC_API_ORIGIN)" npm run dev

stop-web-next$(1):
	@pids="$$$$(lsof -tiTCP:$$(GC_WEB_NEXT$(1)_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$$$pids" ]; then \
		echo "Stopping gc-web-next$(1) on port $$(GC_WEB_NEXT$(1)_PORT): $$$$pids"; \
		kill -9 $$$$pids; \
		while lsof -tiTCP:$$(GC_WEB_NEXT$(1)_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "gc-web-next$(1) port $$(GC_WEB_NEXT$(1)_PORT) is not in use"; \
	fi
endef

define CHECK_WEB_NEXT_TARGET
check-web-next$(1):
	@cd "$$(GC_WEB_NEXT$(1)_DIR)" && npm run check
endef

$(eval $(call RUN_WEB_NEXT_TARGET,01))
$(eval $(call RUN_WEB_NEXT_TARGET,02))
$(eval $(call RUN_WEB_NEXT_TARGET,03))
$(eval $(call RUN_WEB_NEXT_TARGET,04))
$(eval $(call RUN_WEB_NEXT_TARGET,05))
$(eval $(call CHECK_WEB_NEXT_TARGET,02))
$(eval $(call CHECK_WEB_NEXT_TARGET,03))
$(eval $(call CHECK_WEB_NEXT_TARGET,04))
$(eval $(call CHECK_WEB_NEXT_TARGET,05))

run-web-next-all: stop-web-next-all
	@echo "Starting gc-web-next02..05 Vue frontends"
	@set -e; \
	(cd "$(GC_WEB_NEXT02_DIR)" && PORT="$(GC_WEB_NEXT02_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev) & p02=$$!; \
	(cd "$(GC_WEB_NEXT03_DIR)" && PORT="$(GC_WEB_NEXT03_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev) & p03=$$!; \
	(cd "$(GC_WEB_NEXT04_DIR)" && PORT="$(GC_WEB_NEXT04_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev) & p04=$$!; \
	(cd "$(GC_WEB_NEXT05_DIR)" && PORT="$(GC_WEB_NEXT05_PORT)" V3_API_ORIGIN="$(GC_API_ORIGIN)" npm run dev) & p05=$$!; \
	trap 'kill -9 $$p02 $$p03 $$p04 $$p05 2>/dev/null || true' INT TERM EXIT; \
	wait $$p02 $$p03 $$p04 $$p05

stop-web-next-all: stop-web-next02 stop-web-next03 stop-web-next04 stop-web-next05

check-web-next-all: check-web-next02 check-web-next03 check-web-next04 check-web-next05

run-next: run-web-next-all

stop-next: stop-web-next-all

check-next: check-web-next-all

endif
