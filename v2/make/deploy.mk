# =========================================
# make/deploy.mk
# Public deploy entrypoints
# =========================================

.PHONY: deploy-help docker-config docker-build docker-up docker-down docker-logs deploy-ps deploy-logs

ROOT_DIR ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/local.secrets.mk

DEPLOY_LOCAL_DIR ?= $(ROOT_DIR)/deploy
COMPOSE ?= docker-compose
GC_API_IMAGE ?= rudy-gc-v2-api
GC_API_TAG ?= latest
GC_WEB_IMAGE ?= rudy-gc-v2-web
GC_WEB_TAG ?= latest

COMPOSE_ENV = GC_API_IMAGE="$(GC_API_IMAGE)" GC_API_TAG="$(GC_API_TAG)" GC_WEB_IMAGE="$(GC_WEB_IMAGE)" GC_WEB_TAG="$(GC_WEB_TAG)"

deploy-help:
	@echo "Deploy targets:"
	@echo "  make docker-config     Check local compose config"
	@echo "  make docker-build      Build Rudy GC API and web images"
	@echo "  make docker-up         Start Rudy GC v2 compose services"
	@echo "  make docker-down       Stop Rudy GC v2 compose services"
	@echo "  make docker-logs       Follow Rudy GC v2 compose logs"
	@echo "  make deploy-ps         Show compose services"
	@echo "  make deploy-logs       Show compose logs"
	@echo "  make deploy-help       Show deploy targets"
	@echo ""
	@echo "Deploy variables:"
	@echo "  DEPLOY_LOCAL_DIR=$(DEPLOY_LOCAL_DIR)"
	@echo "  COMPOSE=$(COMPOSE)"
	@echo "  GC_API_IMAGE=$(GC_API_IMAGE):$(GC_API_TAG)"
	@echo "  GC_WEB_IMAGE=$(GC_WEB_IMAGE):$(GC_WEB_TAG)"

docker-config:
	@cd "$(DEPLOY_LOCAL_DIR)" && $(COMPOSE_ENV) $(COMPOSE) config

docker-build:
	@cd "$(DEPLOY_LOCAL_DIR)" && $(COMPOSE_ENV) $(COMPOSE) build

docker-up:
	@cd "$(DEPLOY_LOCAL_DIR)" && $(COMPOSE_ENV) $(COMPOSE) up -d

docker-down:
	@cd "$(DEPLOY_LOCAL_DIR)" && $(COMPOSE_ENV) $(COMPOSE) down

docker-logs:
	@cd "$(DEPLOY_LOCAL_DIR)" && $(COMPOSE_ENV) $(COMPOSE) logs -f

deploy-ps:
	@cd "$(DEPLOY_LOCAL_DIR)" && $(COMPOSE_ENV) $(COMPOSE) ps

deploy-logs: docker-logs
