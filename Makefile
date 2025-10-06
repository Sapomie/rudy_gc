APP := rudy_gc

# ===== 可覆盖的参数（执行 make 时可用 ENV 覆盖） =====
DB_URL ?= "root:4521822123@tcp(127.0.0.1:3306)/rudy_gc"
GOCTL  ?= goctl
STYLE  ?= go_zero        # go_zero / snake / go_zero

MOVIE_MODEL_DIR  := data/modelx/moviex
SPIDER_MODEL_DIR := data/modelx/spiderx

# 分类规则
MOVIE_TABLES := "a*,am*,bm*,c*,e*,v*"
SPIDER_TABLES := "d*"

.PHONY: run lint gen-model-movie gen-model-spider gen-model clean-model migrate auto-migrate docker-build

run:
	go run ./cmd/api

lint:
	@echo "Add golangci-lint if needed"

# 生成 moviex 表（带缓存）
gen-model-movie:
	$(GOCTL) model mysql datasource \
		-url=$(DB_URL) \
		-table=$(MOVIE_TABLES) \
		-dir=$(MOVIE_MODEL_DIR) \
		--style=$(STYLE) \
		-c

# 生成 spiderx 表（不带缓存）
gen-model-spider:
	$(GOCTL) model mysql datasource \
		-url=$(DB_URL) \
		-table=$(SPIDER_TABLES) \
		-dir=$(SPIDER_MODEL_DIR) \
		--style=$(STYLE)

# 一起生成
gen-model: gen-model-movie gen-model-spider

clean-model:
	rm -rf $(MOVIE_MODEL_DIR) $(SPIDER_MODEL_DIR)

migrate:
	bash ./scripts/migrate.sh

auto-migrate:
	go run ./cmd/db_migrate

docker-build:
	docker build -f deploy/docker/api.Dockerfile -t $(APP)/api:dev .

# ===== Redis 参数（可通过 make 传参覆盖） =====
REDIS_HOST ?= 127.0.0.1
REDIS_PORT ?= 6379
REDIS_DB   ?= 0

# ===== 清理所有 rudy_gc 相关缓存 =====
clear-cache:
	@COUNT=$$(redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) KEYS "cache:rudyGc:*" | wc -l); \
	echo "⚡ Found $$COUNT rudy_gc cache keys."; \
	if [ $$COUNT -gt 0 ]; then \
		redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) KEYS "cache:rudyGc:*" | xargs -r redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) DEL > /dev/null; \
		echo "✅ Cleared $$COUNT keys."; \
	else \
		echo "ℹ️  Nothing to clear."; \
	fi