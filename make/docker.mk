# make/docker.mk

ROOT_DIR ?= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

.PHONY: docker-build

docker-build:
	docker build -f deploy/docker/api.Dockerfile -t $(APP)/api:dev .