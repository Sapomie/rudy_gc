# make/migrate.mk
LEGACY_ROOT ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
-include $(LEGACY_ROOT)/make/common.mk


.PHONY: migrate auto-migrate


auto-migrate:
	cd "$(ROOT_DIR)" && go run ./cmd/db_migrate
