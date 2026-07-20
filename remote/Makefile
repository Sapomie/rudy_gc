ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/apps.mk
include $(MAKE_DIR)/verify.mk

.PHONY: help

help:
	@echo "Remote project targets:"
	@echo "  make run-api      Run remote API on $(API_PORT)"
	@echo "  make run-web      Run remote web on $(WEB_PORT)"
	@echo "  make run-dev      Run remote API and web"
	@echo "  make stop-all     Stop remote API and web"
	@echo "  make migrate-api  Migrate remote database"
	@echo "  make verify       Verify remote API and web"
