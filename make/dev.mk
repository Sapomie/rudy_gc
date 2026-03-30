# make/dev.mk
ROOT_DIR ?= /Users/gaojinwei/src/my/rudy_gc
API_CONFIG_5010 := $(ROOT_DIR)/cmd/api/config.yaml
API_CONFIG_5011 := /tmp/rudy_gc_api_5011.yaml

.PHONY: run run-5011 lint
run:
	cd "$(ROOT_DIR)" && go run ./cmd/api

run-5011:
	cp "$(API_CONFIG_5010)" "$(API_CONFIG_5011)"
	perl -pi -e 's/Port:\s*":5010"/Port: ":5011"/' "$(API_CONFIG_5011)"
	@echo "[run-5011] config=$(API_CONFIG_5011)"
	cd "$(ROOT_DIR)" && go run ./cmd/api -f "$(API_CONFIG_5011)"

lint:
	@echo "Add golangci-lint if needed"
