# make/common.mk
APP ?= rudy_gc

# ===== Go / DB / Redis 默认配置 =====
DB_URL     ?= root:4521822123@tcp(127.0.0.1:3306)/rudy_gc
GOCTL      ?= goctl
STYLE      ?= go_zero
REDIS_HOST ?= 127.0.0.1
REDIS_PORT ?= 6379
REDIS_DB   ?= 0

MOVIE_MODEL_DIR  ?= data/modelx/moviex
SPIDER_MODEL_DIR ?= data/modelx/spiderx
MOVIE_TABLES     ?= "a*,am*,bm*,c*,e*,v*"
SPIDER_TABLES    ?= "d*"

.PHONY: help
help:
	@echo "Available make modules:"
	@echo "  dev.mk     - run / lint"
	@echo "  model.mk   - generate goctl model"
	@echo "  migrate.mk - db migrations"
	@echo "  docker.mk  - docker build"
	@echo "  cache.mk   - clear redis cache"