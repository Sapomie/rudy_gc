ifndef REMOTE_PROJECT_VERIFY_MK
REMOTE_PROJECT_VERIFY_MK := 1

REMOTE_VERIFY_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
REMOTE_VERIFY_MAKE_DIR := $(dir $(REMOTE_VERIFY_FILE))

include $(REMOTE_VERIFY_MAKE_DIR)apps.mk

.PHONY: verify

verify: test-api build-api check-web test-web build-web

endif
