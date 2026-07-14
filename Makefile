ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/apps.mk
include $(MAKE_DIR)/verify.mk
include $(MAKE_DIR)/backup.mk
include $(MAKE_DIR)/docker.mk

.PHONY: help

help:
	@echo "Project targets:"
	@echo "  make paths       Show resolved paths and ports"
	@echo "  make run-api     Run API on $(API_PORT)"
	@echo "  make run-web     Run web on $(WEB_PORT)"
	@echo "  make run-dev     Run API and web together"
	@echo "  make stop-api    Stop API port"
	@echo "  make stop-web    Stop web port"
	@echo "  make stop-all    Stop API and web ports"
	@echo "  make test-api    Run Go tests"
	@echo "  make build-api   Build Go packages"
	@echo "  make migrate-api Run database migrations"
	@echo "  make check-web   Type-check web"
	@echo "  make test-web    Run web unit tests"
	@echo "  make build-web   Build web"
	@echo "  make verify      Run API tests and web checks/build"
	@echo "  make db-backup   Backup local rudy_gc MySQL database"
	@echo "  make backup      Alias of db-backup"
	@echo "  make docker-build      Build Docker images"
	@echo "  make docker-import-db  Import host MySQL into Docker MySQL"
	@echo "  make docker-up         Start Docker services"
	@echo "  make docker-down       Stop Docker services"
	@echo "  make docker-rebuild    Build images, import DB, then start services"
