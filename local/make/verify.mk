ifndef PROJECT_VERIFY_MK
PROJECT_VERIFY_MK := 1

PROJECT_VERIFY_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_VERIFY_MAKE_DIR := $(dir $(PROJECT_VERIFY_FILE))

include $(PROJECT_VERIFY_MAKE_DIR)common.mk
include $(PROJECT_VERIFY_MAKE_DIR)apps.mk

.PHONY: verify build-all test-all check-dev

test-all: test-api

build-all: build-api build-web

verify: test-api build-api check-web test-web build-web

endif
