ifndef PROJECT_COMMON_MK
PROJECT_COMMON_MK := 1

PROJECT_COMMON_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_MAKE_DIR := $(dir $(PROJECT_COMMON_FILE))

ROOT_DIR ?= $(abspath $(PROJECT_MAKE_DIR)/..)
API_DIR ?= $(ROOT_DIR)/api
WEB_DIR ?= $(ROOT_DIR)/web

API_CONFIG_FILE ?= $(API_DIR)/config/config.yaml
API_PORT ?= 5021
WEB_PORT ?= 5022
API_ORIGIN ?= http://127.0.0.1:$(API_PORT)

.PHONY: paths

paths:
	@echo "ROOT_DIR=$(ROOT_DIR)"
	@echo "API_DIR=$(API_DIR)"
	@echo "WEB_DIR=$(WEB_DIR)"
	@echo "API_CONFIG_FILE=$(API_CONFIG_FILE)"
	@echo "API_PORT=$(API_PORT)"
	@echo "WEB_PORT=$(WEB_PORT)"
	@echo "API_ORIGIN=$(API_ORIGIN)"

endif
