APP := rudy_gc

# ===== 可覆盖的参数（执行 make 时可用 ENV 覆盖） =====
DB_URL    ?= "root:4521822123@tcp(127.0.0.1:3306)/rudy_gc"
MODEL_DIR ?= data/modelx/spider
TABLES    ?= "*"            # 也可以指定 "user,order" 逗号分隔
GOCTL     ?= goctl
STYLE     ?= go_zero        # go_zero / snake / go_zero

.PHONY: run lint gen-model clean-model migrate docker-build

run:
	go run ./cmd/api

lint:
	@echo "Add golangci-lint if needed"

# 生成 goctl model（支持通配和多表），-c 开启 custom 扩展
gen-model:
	$(GOCTL) model mysql datasource \
		-url=$(DB_URL) \
		-table=$(TABLES) \
		-dir=$(MODEL_DIR) \
		--style=$(STYLE) \
#		-c

# 清理再生成（可选）
clean-model:
	rm -rf $(MODEL_DIR)

migrate:
	bash ./scripts/migrate.sh

docker-build:
	docker build -f deploy/docker/api.Dockerfile -t $(APP)/api:dev .
