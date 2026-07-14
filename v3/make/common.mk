ifndef V3_COMMON_MK
V3_COMMON_MK := 1

V3_COMMON_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_MAKE_DIR := $(dir $(V3_COMMON_FILE))

ROOT_DIR ?= $(abspath $(V3_MAKE_DIR)/..)
GC_API_DIR ?= $(ROOT_DIR)/gc-api
DEPLOY_DIR ?= $(ROOT_DIR)/deploy

GC_API_PORT ?= 5013
GC_API_ORIGIN ?= http://127.0.0.1:$(GC_API_PORT)

.PHONY: paths

paths:
	@echo "ROOT_DIR=$(ROOT_DIR)"
	@echo "GC_API_DIR=$(GC_API_DIR)"
	@echo "GC_WEB_NEXT02_DIR=$(GC_WEB_NEXT02_DIR)"
	@echo "DEPLOY_DIR=$(DEPLOY_DIR)"
	@echo "GC_API_PORT=$(GC_API_PORT)"
	@echo "GC_WEB_NEXT02_PORT=$(GC_WEB_NEXT02_PORT)"
	@echo "GC_API_ORIGIN=$(GC_API_ORIGIN)"

endif
