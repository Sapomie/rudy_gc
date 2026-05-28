ifndef V3_COMMON_MK
V3_COMMON_MK := 1

V3_COMMON_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_MAKE_DIR := $(dir $(V3_COMMON_FILE))

ROOT_DIR ?= $(abspath $(V3_MAKE_DIR)/..)
GC_API_DIR ?= $(ROOT_DIR)/gc-api
GC_WEB_DIR ?= $(ROOT_DIR)/gc-web
DEPLOY_DIR ?= $(ROOT_DIR)/deploy

GC_API_PORT ?= 5013
GC_WEB_PORT ?= 5012
GC_API_ORIGIN ?= http://127.0.0.1:$(GC_API_PORT)
WEB_HOST ?= 127.0.0.1

.PHONY: paths

paths:
	@echo "ROOT_DIR=$(ROOT_DIR)"
	@echo "GC_API_DIR=$(GC_API_DIR)"
	@echo "GC_WEB_DIR=$(GC_WEB_DIR)"
	@echo "DEPLOY_DIR=$(DEPLOY_DIR)"
	@echo "GC_API_PORT=$(GC_API_PORT)"
	@echo "GC_WEB_PORT=$(GC_WEB_PORT)"
	@echo "GC_API_ORIGIN=$(GC_API_ORIGIN)"
	@echo "WEB_HOST=$(WEB_HOST)"

endif
