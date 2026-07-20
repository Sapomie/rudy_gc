ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/apps.mk
include $(MAKE_DIR)/verify.mk
include $(MAKE_DIR)/backup.mk

.PHONY: help

help:
	@echo "Local project targets:"
	@echo "  make run-api      Run local API on $(API_PORT)"
	@echo "  make run-web      Run local web on $(WEB_PORT)"
	@echo "  make run-dev      Run local API and web"
	@echo "  make stop-all     Stop local API and web"
	@echo "  make migrate-api  Migrate local database"
	@echo "  make verify       Verify local API and web"
	@echo "  make backup       Backup local database"
