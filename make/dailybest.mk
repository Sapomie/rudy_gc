ifndef PROJECT_DAILYBEST_MK
PROJECT_DAILYBEST_MK := 1

PROJECT_DAILYBEST_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DAILYBEST_MAKE_DIR := $(dir $(PROJECT_DAILYBEST_FILE))

include $(PROJECT_DAILYBEST_MAKE_DIR)common.mk

DAILYBEST_REMOTE_CONFIG_FILE ?= $(API_DIR)/cmd/dailybest-remote/config.yaml
DAILYBEST_REMOTE_PORT ?= 8891

.PHONY: dailybest-help run-dailybest-remote stop-dailybest-remote \
	build-dailybest-remote test-dailybest-remote migrate-dailybest-remote

dailybest-help:
	@echo "DailyBest remote targets:"
	@echo "  make run-dailybest-remote      Run the remote collector and API"
	@echo "  make stop-dailybest-remote     Stop the remote API port"
	@echo "  make build-dailybest-remote    Build the remote collector and API"
	@echo "  make test-dailybest-remote     Test DailyBest remote and sync packages"
	@echo "  make migrate-dailybest-remote  Migrate the remote DailyBest database"
	@echo "  DAILYBEST_REMOTE_CONFIG_FILE=$(DAILYBEST_REMOTE_CONFIG_FILE)"
	@echo "  DAILYBEST_REMOTE_PORT=$(DAILYBEST_REMOTE_PORT)"

run-dailybest-remote: stop-dailybest-remote
	@cd "$(API_DIR)" && go run ./cmd/dailybest-remote -f "$(DAILYBEST_REMOTE_CONFIG_FILE)"

stop-dailybest-remote:
	@pids="$$(lsof -tiTCP:$(DAILYBEST_REMOTE_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping DailyBest remote service on port $(DAILYBEST_REMOTE_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(DAILYBEST_REMOTE_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "DailyBest remote port is not in use"; \
	fi

build-dailybest-remote:
	@cd "$(API_DIR)" && go build ./cmd/dailybest-remote ./cmd/dailybest-remote-migrate

test-dailybest-remote:
	@cd "$(API_DIR)" && go test ./internal/contracts ./internal/domain/dailybest \
		./internal/repository/dailybestremotex ./internal/service/dailybestcollector \
		./internal/service/dailybestsync ./internal/platform/dailybestapi \
		./internal/transport/http/dailybestremote

migrate-dailybest-remote:
	@cd "$(API_DIR)" && go run ./cmd/dailybest-remote-migrate -f "$(DAILYBEST_REMOTE_CONFIG_FILE)"

endif
