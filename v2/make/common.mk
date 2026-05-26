ROOT_DIR ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
GC_API_DIR ?= $(ROOT_DIR)/gc-api
GC_WEB_DIR ?= $(ROOT_DIR)/gc-web
DEPLOY_LOCAL_DIR ?= $(ROOT_DIR)/deploy
DEPLOY_DIR ?= $(DEPLOY_LOCAL_DIR)

paths:
	@echo "ROOT_DIR=$(ROOT_DIR)"
	@echo "GC_API_DIR=$(GC_API_DIR)"
	@echo "GC_WEB_DIR=$(GC_WEB_DIR)"
	@echo "DEPLOY_LOCAL_DIR=$(DEPLOY_LOCAL_DIR)"
