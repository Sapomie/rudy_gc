ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/gc-api.mk
include $(MAKE_DIR)/gc-web.mk
include $(MAKE_DIR)/run.mk
include $(MAKE_DIR)/verify.mk

.PHONY: help

help:
	@echo "v3 legacy split targets:"
	@echo "  make paths       Show resolved paths and ports"
	@echo "  make run-api     Run v3 API on $(GC_API_PORT)"
	@echo "  make run-web     Run v3 web on $(GC_WEB_PORT)"
	@echo "  make run-all     Run v3 API and web together"
	@echo "  make stop-api    Stop v3 API port"
	@echo "  make stop-web    Stop v3 web port"
	@echo "  make stop-all    Stop v3 API and web ports"
	@echo "  make test-api    Run Go tests"
	@echo "  make build-api   Build Go packages"
	@echo "  make check-web   Run web syntax checks"
	@echo "  make verify      Run API tests and web checks"
