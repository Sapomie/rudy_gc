ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/apps.mk
include $(MAKE_DIR)/dailybest.mk
include $(MAKE_DIR)/verify.mk
include $(MAKE_DIR)/backup.mk

.PHONY: help

help:
	@echo "Project targets:"
	@echo "  make paths       Show resolved paths and ports"
	@echo "  make run-api     Run API on $(API_PORT)"
	@echo "  make run-web     Run web on $(WEB_PORT)"
	@echo "  make run-dev     Run API and web together"
	@echo "  make run-all     Run API and web together"
	@echo "  make stop-api    Stop API port"
	@echo "  make stop-web    Stop web port"
	@echo "  make stop-all    Stop API and web ports"
	@echo "  make test-api    Run Go tests"
	@echo "  make build-api   Build Go packages"
	@echo "  make migrate-api Run database migrations"
	@echo "  make run-dailybest-remote     Run DailyBest remote service"
	@echo "  make build-dailybest-remote   Build DailyBest remote service"
	@echo "  make migrate-dailybest-remote Migrate DailyBest remote database"
	@echo "  make check-web   Type-check web"
	@echo "  make test-web    Run web unit tests"
	@echo "  make build-web   Build web"
	@echo "  make verify      Run API tests and web checks/build"
	@echo "  make db-backup   Backup local rudy_gc MySQL database"
	@echo "  make backup      Alias of db-backup"
