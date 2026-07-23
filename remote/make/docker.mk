ifndef REMOTE_PROJECT_DOCKER_MK
REMOTE_PROJECT_DOCKER_MK := 1

REMOTE_DOCKER_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
REMOTE_DOCKER_MAKE_DIR := $(dir $(REMOTE_DOCKER_FILE))

include $(REMOTE_DOCKER_MAKE_DIR)common.mk

DOCKER_COMPOSE ?= docker-compose
DOCKER_COMPOSE_FILE ?= $(ROOT_DIR)/deploy/docker/compose.yaml

.PHONY: docker-config docker-build docker-migrate docker-up docker-down docker-status docker-logs

docker-config:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" config

docker-build:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" build

docker-migrate:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" run --rm --no-deps --entrypoint /app/remote-migrate remote-api -f /app/config/server.yaml

docker-up:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" up -d

docker-down:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" down

docker-status:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" ps

docker-logs:
	@$(DOCKER_COMPOSE) -f "$(DOCKER_COMPOSE_FILE)" logs

endif
