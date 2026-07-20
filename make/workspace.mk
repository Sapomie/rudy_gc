ifndef RUDY_GC_WORKSPACE_MK
RUDY_GC_WORKSPACE_MK := 1

WORKSPACE_MAKE_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
WORKSPACE_ROOT := $(abspath $(dir $(WORKSPACE_MAKE_FILE))/..)
LOCAL_DIR := $(WORKSPACE_ROOT)/local
REMOTE_DIR := $(WORKSPACE_ROOT)/remote
SHARED_GO_DIR := $(WORKSPACE_ROOT)/shared/go

.PHONY: local-help remote-help run-local run-remote stop-local stop-remote \
	verify-local verify-remote verify-shared verify

local-help:
	@$(MAKE) -C "$(LOCAL_DIR)" help

remote-help:
	@$(MAKE) -C "$(REMOTE_DIR)" help

run-local:
	@$(MAKE) -C "$(LOCAL_DIR)" run-dev

run-remote:
	@$(MAKE) -C "$(REMOTE_DIR)" run-dev

stop-local:
	@$(MAKE) -C "$(LOCAL_DIR)" stop-all

stop-remote:
	@$(MAKE) -C "$(REMOTE_DIR)" stop-all

verify-local:
	@$(MAKE) -C "$(LOCAL_DIR)" verify

verify-remote:
	@$(MAKE) -C "$(REMOTE_DIR)" verify

verify-shared:
	@cd "$(SHARED_GO_DIR)" && go test ./...

verify: verify-shared verify-local verify-remote

endif
