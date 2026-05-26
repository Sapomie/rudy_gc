# make/docker.mk

LEGACY_ROOT ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
-include $(LEGACY_ROOT)/make/common.mk

.PHONY: docker-build

docker-build:
	docker build -f "$(LEGACY_ROOT)/deploy/docker/api.Dockerfile" -t $(APP)/api:dev "$(ROOT_DIR)"
