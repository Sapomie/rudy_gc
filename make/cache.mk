# make/cache.mk
ROOT_DIR ?= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

# 默认变量（如果 common.mk 没加载）
REDIS_HOST ?= 127.0.0.1
REDIS_PORT ?= 6379
REDIS_DB   ?= 0
BIZ_REDIS_HOST ?= 127.0.0.1
BIZ_REDIS_PORT ?= 6378
BIZ_REDIS_DB   ?= 0

.PHONY: clear-cache
clear-cache:
	@CORE_COUNT=$$(redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) KEYS "cache:rudyGc:*" | wc -l); \
	echo "⚡ [core] Found $$CORE_COUNT keys (cache:rudyGc:* @ $(REDIS_HOST):$(REDIS_PORT)/$(REDIS_DB))."; \
	if [ $$CORE_COUNT -gt 0 ]; then \
		redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) KEYS "cache:rudyGc:*" | xargs -r redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) DEL > /dev/null; \
		echo "✅ [core] Cleared $$CORE_COUNT keys."; \
	else \
		echo "ℹ️  [core] Nothing to clear."; \
	fi; \
	BIZ_COUNT=$$(redis-cli -h $(BIZ_REDIS_HOST) -p $(BIZ_REDIS_PORT) -n $(BIZ_REDIS_DB) KEYS "rudy:mt:*" | wc -l); \
	echo "⚡ [biz] Found $$BIZ_COUNT keys (rudy:mt:* @ $(BIZ_REDIS_HOST):$(BIZ_REDIS_PORT)/$(BIZ_REDIS_DB))."; \
	if [ $$BIZ_COUNT -gt 0 ]; then \
		redis-cli -h $(BIZ_REDIS_HOST) -p $(BIZ_REDIS_PORT) -n $(BIZ_REDIS_DB) KEYS "rudy:mt:*" | xargs -r redis-cli -h $(BIZ_REDIS_HOST) -p $(BIZ_REDIS_PORT) -n $(BIZ_REDIS_DB) DEL > /dev/null; \
		echo "✅ [biz] Cleared $$BIZ_COUNT keys."; \
	else \
		echo "ℹ️  [biz] Nothing to clear."; \
	fi
