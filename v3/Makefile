ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/common.mk
include $(MAKE_DIR)/gc-api.mk
include $(MAKE_DIR)/gc-web.mk
include $(MAKE_DIR)/gc-web-next.mk
include $(MAKE_DIR)/run.mk
include $(MAKE_DIR)/verify.mk
include $(MAKE_DIR)/backup.mk
include $(MAKE_DIR)/docker.mk

.PHONY: help

help:
	@echo "v3 legacy split targets:"
	@echo "  make paths       Show resolved paths and ports"
	@echo "  make run-api     Run v3 API on $(GC_API_PORT)"
	@echo "  make run-web     Run v3 web on $(GC_WEB_PORT)"
	@echo "  make run-web-next01..05  Run v3 next frontends on ports 5021..5025"
	@echo "  make check-web-next-all  Check v3 next02..05 frontends"
	@echo "  make run-next    Run all v3 next frontends"
	@echo "  make stop-next   Stop all v3 next frontends"
	@echo "  make check-next  Check v3 next02..05 frontends"
	@echo "  make run-all     Run v3 API and web together"
	@echo "  make stop-api    Stop v3 API port"
	@echo "  make stop-web    Stop v3 web port"
	@echo "  make stop-all    Stop v3 API and web ports"
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
