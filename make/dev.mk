# make/dev.mk
.PHONY: run lint
run:
	go run ./cmd/api

lint:
	@echo "Add golangci-lint if needed"