# make/model.mk
ROOT_DIR ?= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

.PHONY: gen-model gen-model-movie gen-model-spider clean-model

gen-model-movie:
	$(GOCTL) model mysql datasource \
		-url=$(DB_URL) \
		-table=$(MOVIE_TABLES) \
		-dir=$(MOVIE_MODEL_DIR) \
		--style=$(STYLE) \
		-c

gen-model-spider:
	$(GOCTL) model mysql datasource \
		-url=$(DB_URL) \
		-table=$(SPIDER_TABLES) \
		-dir=$(SPIDER_MODEL_DIR) \
		--style=$(STYLE)

gen-model: gen-model-movie gen-model-spider

clean-model:
	rm -rf $(MOVIE_MODEL_DIR) $(SPIDER_MODEL_DIR)