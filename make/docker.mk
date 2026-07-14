ifndef PROJECT_DOCKER_MK
PROJECT_DOCKER_MK := 1

PROJECT_DOCKER_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DOCKER_MAKE_DIR := $(dir $(PROJECT_DOCKER_FILE))

include $(PROJECT_DOCKER_MAKE_DIR)common.mk

DOCKER_COMPOSE ?= docker compose
DOCKER_DIR ?= $(DEPLOY_DIR)/docker
DOCKER_COMPOSE_FILE ?= $(DOCKER_DIR)/compose.yml
DOCKER_PROJECT_NAME ?= rudy_gc_v3
DOCKER_IMPORT_DB_SCRIPT ?= $(DOCKER_DIR)/scripts/import_host_mysql_to_docker.sh
DOCKER_PROXY ?= http://127.0.0.1:7897/
DOCKER_ENV := HTTP_PROXY="$(DOCKER_PROXY)" HTTPS_PROXY="$(DOCKER_PROXY)" ALL_PROXY="$(DOCKER_PROXY)" http_proxy="$(DOCKER_PROXY)" https_proxy="$(DOCKER_PROXY)" all_proxy="$(DOCKER_PROXY)"

.PHONY: docker-help docker-build docker-import-db docker-up docker-down docker-logs docker-ps docker-rebuild

docker-help:
	@echo "Docker targets:"
	@echo "  make docker-build      Build v3 API and web images"
	@echo "  make docker-import-db  Import host MySQL into Docker MySQL"
	@echo "  make docker-up         Start v3 API and web containers"
	@echo "  make docker-down       Stop v3 API and web containers"
	@echo "  make docker-logs       Follow v3 container logs"
	@echo "  make docker-ps         Show v3 compose status"
	@echo "  make docker-rebuild    Build images, import DB, then start services"
	@echo "Docker variables:"
	@echo "  DOCKER_COMPOSE_FILE=$(DOCKER_COMPOSE_FILE)"
	@echo "  DOCKER_PROJECT_NAME=$(DOCKER_PROJECT_NAME)"
	@echo "  DOCKER_IMPORT_DB_SCRIPT=$(DOCKER_IMPORT_DB_SCRIPT)"
	@echo "  DOCKER_PROXY=$(DOCKER_PROXY)"

docker-build:
	@cd "$(ROOT_DIR)" && $(DOCKER_ENV) $(DOCKER_COMPOSE) -p "$(DOCKER_PROJECT_NAME)" -f "$(DOCKER_COMPOSE_FILE)" build

docker-import-db:
	@"$(DOCKER_IMPORT_DB_SCRIPT)"

docker-up:
	@cd "$(ROOT_DIR)" && $(DOCKER_ENV) $(DOCKER_COMPOSE) -p "$(DOCKER_PROJECT_NAME)" -f "$(DOCKER_COMPOSE_FILE)" up -d

docker-down:
	@cd "$(ROOT_DIR)" && $(DOCKER_ENV) $(DOCKER_COMPOSE) -p "$(DOCKER_PROJECT_NAME)" -f "$(DOCKER_COMPOSE_FILE)" down

docker-logs:
	@cd "$(ROOT_DIR)" && $(DOCKER_ENV) $(DOCKER_COMPOSE) -p "$(DOCKER_PROJECT_NAME)" -f "$(DOCKER_COMPOSE_FILE)" logs -f

docker-ps:
	@cd "$(ROOT_DIR)" && $(DOCKER_ENV) $(DOCKER_COMPOSE) -p "$(DOCKER_PROJECT_NAME)" -f "$(DOCKER_COMPOSE_FILE)" ps

docker-rebuild: docker-build docker-import-db docker-up

endif
