ifndef REMOTE_PROJECT_DOCKER_MK
REMOTE_PROJECT_DOCKER_MK := 1

REMOTE_DOCKER_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
REMOTE_DOCKER_MAKE_DIR := $(dir $(REMOTE_DOCKER_FILE))

include $(REMOTE_DOCKER_MAKE_DIR)common.mk

DOCKER_COMPOSE ?= docker compose
DOCKER_COMPOSE_FILE ?= $(ROOT_DIR)/deploy/docker/compose.yaml
REMOTE_ENV_FILE ?= $(ROOT_DIR)/deploy/docker/remote-release.env
REMOTE_CONFIG_FILE ?= /var/snap/docker/common/rudy_gc/remote/config/server.yaml
REMOTE_NGINX_CONFIG_FILE ?= $(ROOT_DIR)/deploy/docker/nginx.conf
RELEASE_TAG ?=
OFFLINE_OUTPUT_DIR ?= $(abspath $(ROOT_DIR)/../tmp/remote-offline/$(RELEASE_TAG))
BUILD_OFFLINE_SCRIPT ?= $(abspath $(ROOT_DIR)/../scripts/build-remote-offline-bundle.sh)
INSTALL_OFFLINE_SCRIPT ?= $(abspath $(ROOT_DIR)/../scripts/install-remote-offline.sh)
DOCKER_COMPOSE_ARGS := -p rudy-gc-remote --env-file "$(REMOTE_ENV_FILE)" -f "$(DOCKER_COMPOSE_FILE)"

.PHONY: docker-config docker-build docker-offline-bundle docker-offline-install \
	docker-migrate docker-up docker-down docker-status docker-logs

docker-config:
	@REMOTE_CONFIG_FILE="$(REMOTE_CONFIG_FILE)" REMOTE_NGINX_CONFIG_FILE="$(REMOTE_NGINX_CONFIG_FILE)" $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS) config

docker-build: docker-offline-bundle

docker-offline-bundle:
	@test -n "$(RELEASE_TAG)" || { echo "RELEASE_TAG is required" >&2; exit 1; }
	@"$(BUILD_OFFLINE_SCRIPT)" "$(RELEASE_TAG)" "$(OFFLINE_OUTPUT_DIR)"

docker-offline-install:
	@test -n "$(BUNDLE_DIR)" || { echo "BUNDLE_DIR is required" >&2; exit 1; }
	@"$(INSTALL_OFFLINE_SCRIPT)" "$(BUNDLE_DIR)" "$(REMOTE_CONFIG_FILE)"

docker-migrate:
	@REMOTE_CONFIG_FILE="$(REMOTE_CONFIG_FILE)" REMOTE_NGINX_CONFIG_FILE="$(REMOTE_NGINX_CONFIG_FILE)" $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS) run --rm --no-deps --pull never --entrypoint /app/remote-migrate remote-api -f /app/config/server.yaml

docker-up:
	@REMOTE_CONFIG_FILE="$(REMOTE_CONFIG_FILE)" REMOTE_NGINX_CONFIG_FILE="$(REMOTE_NGINX_CONFIG_FILE)" $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS) up -d --pull never --no-build

docker-down:
	@REMOTE_CONFIG_FILE="$(REMOTE_CONFIG_FILE)" REMOTE_NGINX_CONFIG_FILE="$(REMOTE_NGINX_CONFIG_FILE)" $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS) down

docker-status:
	@REMOTE_CONFIG_FILE="$(REMOTE_CONFIG_FILE)" REMOTE_NGINX_CONFIG_FILE="$(REMOTE_NGINX_CONFIG_FILE)" $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS) ps

docker-logs:
	@REMOTE_CONFIG_FILE="$(REMOTE_CONFIG_FILE)" REMOTE_NGINX_CONFIG_FILE="$(REMOTE_NGINX_CONFIG_FILE)" $(DOCKER_COMPOSE) $(DOCKER_COMPOSE_ARGS) logs

endif
