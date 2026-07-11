ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/gc-apps.mk
include $(MAKE_DIR)/run.mk
include $(MAKE_DIR)/verify.mk
include $(MAKE_DIR)/backup.mk
include $(MAKE_DIR)/docker.mk

.PHONY: help

help:
	@echo "v3 legacy split targets:"
	@echo "  make paths       Show resolved paths and ports"
	@echo "  make run-api     Run v3 API on $(GC_API_PORT)"
	@echo "  make run-web     Run legacy gc-web on $(GC_WEB_PORT)"
	@echo "  make run-web-next02  Run gc-web-next02 on $(GC_WEB_NEXT02_PORT)"
	@echo "  make stop-web-next02 Stop gc-web-next02 port"
	@echo "  make check-web-next02 Check gc-web-next02"
	@echo "  make run-next    Alias of run-web-next02"
	@echo "  make stop-next   Alias of stop-web-next02"
	@echo "  make check-next  Alias of check-web-next02"
	@echo "  make run-dev     Run API, legacy gc-web, and gc-web-next02 together"
	@echo "  make stop-dev    Stop API, legacy gc-web, and gc-web-next02"
	@echo "  make check-dev   Test API and check both web frontends"
	@echo "  make run-all     Run API, legacy gc-web, and gc-web-next02 together"
	@echo "  make stop-api    Stop v3 API port"
	@echo "  make stop-web    Stop v3 web port"
	@echo "  make stop-all    Stop API, legacy gc-web, and gc-web-next02 ports"
	@echo "  make test-api    Run Go tests"
	@echo "  make build-api   Build Go packages"
	@echo "  make check-web   Run web syntax checks"
	@echo "  make verify      Run API tests and web checks"
	@echo "  make db-backup   Backup local rudy_gc MySQL database"
	@echo "  make backup      Alias of db-backup"
	@echo "  make docker-build      Build v3 Docker images"
	@echo "  make docker-import-db  Import host MySQL into Docker MySQL"
	@echo "  make docker-up         Start v3 Docker services"
	@echo "  make docker-down       Stop v3 Docker services"
	@echo "  make docker-rebuild    Build images, import DB, then start services"
