# make/cache.mk
ROOT_DIR ?= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

# 默认变量（如果 common.mk 没加载）
REDIS_HOST ?= 127.0.0.1
REDIS_PORT ?= 6379
REDIS_DB   ?= 0

.PHONY: clear-cache
clear-cache:
	@COUNT=$$(redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) KEYS "cache:rudyGc:*" | wc -l); \
	echo "⚡ Found $$COUNT cache keys."; \
	if [ $$COUNT -gt 0 ]; then \
		redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) KEYS "cache:rudyGc:*" | xargs -r redis-cli -h $(REDIS_HOST) -p $(REDIS_PORT) -n $(REDIS_DB) DEL > /dev/null; \
		echo "✅ Cleared $$COUNT keys."; \
	else \
		echo "ℹ️  Nothing to clear."; \
	fi