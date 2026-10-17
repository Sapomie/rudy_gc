ifndef RUDY_GC_WORKSPACE_MK
RUDY_GC_WORKSPACE_MK := 1

WORKSPACE_MAKE_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
WORKSPACE_ROOT := $(abspath $(dir $(WORKSPACE_MAKE_FILE))/..)
LOCAL_DIR := $(WORKSPACE_ROOT)/local
REMOTE_DIR := $(WORKSPACE_ROOT)/remote
SHARED_GO_DIR := $(WORKSPACE_ROOT)/shared/go

.PHONY: local-help remote-help run-local run-remote stop-local stop-remote \
	verify-local verify-remote verify-shared verify shared-mysql-bundle

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

shared-mysql-bundle:
	@test -n "$(RELEASE_TAG)" || { echo "RELEASE_TAG is required" >&2; exit 1; }
	@"$(WORKSPACE_ROOT)/scripts/build-shared-mysql-bundle.sh" "$(RELEASE_TAG)" \
		"$(if $(OFFLINE_OUTPUT_DIR),$(OFFLINE_OUTPUT_DIR),$(WORKSPACE_ROOT)/tmp/mysql-offline/$(RELEASE_TAG))"

endif
