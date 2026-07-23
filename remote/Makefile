ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/apps.mk
include $(MAKE_DIR)/verify.mk
include $(MAKE_DIR)/docker.mk

.PHONY: help

help:
	@echo "Remote project targets:"
	@echo "  make run-api      Run remote API on $(API_PORT)"
	@echo "  make run-web      Run remote web on $(WEB_PORT)"
	@echo "  make run-dev      Run remote API and web"
	@echo "  make stop-all     Stop remote API and web"
	@echo "  make migrate-api  Migrate remote database"
	@echo "  make verify       Verify remote API and web"
	@echo "  make docker-config Validate Docker Compose"
	@echo "  make docker-build  Build Remote Docker images"
	@echo "  make docker-migrate Migrate Remote database with the API image"
	@echo "  make docker-up     Start Remote Docker services"
	@echo "  make docker-down   Stop Remote Docker services"
