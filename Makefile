ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
MAKE_DIR := $(ROOT_DIR)/make

include $(MAKE_DIR)/workspace.mk

.PHONY: help

help:
	@echo "Rudy GC workspace targets:"
	@echo "  make local-help     Show local project targets"
	@echo "  make remote-help    Show remote project targets"
	@echo "  make run-local      Run local API and web"
	@echo "  make run-remote     Run remote API and web"
	@echo "  make stop-local     Stop local API and web"
	@echo "  make stop-remote    Stop remote API and web"
	@echo "  make verify-local   Verify local project"
	@echo "  make verify-remote  Verify remote project"
	@echo "  make verify-shared  Verify shared Go modules"
	@echo "  make verify         Verify both projects"
