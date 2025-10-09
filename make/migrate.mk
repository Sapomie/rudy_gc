# make/migrate.mk
ROOT_DIR ?= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk


.PHONY: migrate auto-migrate


auto-migrate:
	go run ../cmd/db_migrate