ifndef V3_VERIFY_MK
V3_VERIFY_MK := 1

V3_VERIFY_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
V3_VERIFY_MAKE_DIR := $(dir $(V3_VERIFY_FILE))

include $(V3_VERIFY_MAKE_DIR)common.mk
include $(V3_VERIFY_MAKE_DIR)gc-apps.mk

.PHONY: verify build-all test-all check-dev

test-all: test-api

build-all: build-api check-web-next02

verify: test-api check-web-next02

endif
