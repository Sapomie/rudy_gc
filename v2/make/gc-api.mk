# =========================================
# make/gc-api.mk
# Rudy GC API local helpers
# =========================================

.PHONY: api-help run-api stop-api test-api build-api gen-model clean-model

ROOT_DIR ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
GC_API_DIR ?= $(ROOT_DIR)/gc-api
GC_GOCTL ?= goctl
GC_MODEL_STYLE ?= go_zero
GC_MODEL_DIR ?= $(GC_API_DIR)/internal/model/modelx
GC_API_CONFIG_FILE ?= $(GC_API_DIR)/config/config.yaml
GC_API_PORT ?= 2041
GC_API_ADDR ?= :$(GC_API_PORT)
GC_DB_URL ?= root:4521822123@tcp(127.0.0.1:3306)/rudy_gc

api-help:
	@echo "Usage examples:"
	@echo "  make run-api"
	@echo "  make stop-api"
	@echo "  make test-api"
	@echo "  make build-api"
	@echo "  make gen-model"
	@echo "  make run-api GC_API_CONFIG_FILE=$(GC_API_CONFIG_FILE)"
	@echo "  make gen-model GC_DB_URL='$(GC_DB_URL)'"

run-api: stop-api
	@echo "Starting Rudy GC API with config: $(GC_API_CONFIG_FILE)"
	@cd "$(GC_API_DIR)" && go run ./cmd/api -f "$(GC_API_CONFIG_FILE)" -addr "$(GC_API_ADDR)"

stop-api:
	@pids="$$(lsof -tiTCP:$(GC_API_PORT) -sTCP:LISTEN)"; \
	if [ -n "$$pids" ]; then \
		echo "Stopping Rudy GC API on port $(GC_API_PORT): $$pids"; \
		kill -9 $$pids; \
		while lsof -tiTCP:$(GC_API_PORT) -sTCP:LISTEN >/dev/null 2>&1; do sleep 0.2; done; \
	else \
		echo "Rudy GC API port $(GC_API_PORT) is not in use"; \
	fi

test-api:
	@cd "$(GC_API_DIR)" && go test ./...

build-api:
	@cd "$(GC_API_DIR)" && go build ./...

gen-model:
	@echo "Generating Rudy GC modelx into: $(GC_MODEL_DIR)"
	@$(GC_GOCTL) model mysql datasource \
		--url="$(GC_DB_URL)" \
		--table="*" \
		--dir="$(GC_MODEL_DIR)" \
		--style="$(GC_MODEL_STYLE)" \
		-c
	@echo "Rudy GC modelx generated: $(GC_MODEL_DIR)"

clean-model:
	@echo "Removing Rudy GC modelx: $(GC_MODEL_DIR)"
	rm -rf "$(GC_MODEL_DIR)"
	@echo "Rudy GC modelx removed."
