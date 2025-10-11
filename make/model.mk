# ================================
# make/model.mk
# 数据模型自动生成（使用 goctl）
# 兼容：无论当前工作目录在哪、是否使用 -C 执行，都写入到项目根下
# ================================

# ---- 项目根路径：基于当前这个 makefile 的真实位置推导 ----
ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

.PHONY: gen-model gen-model-movie gen-model-spider clean-model

# ---- 默认参数（可被外部覆盖）----
DB_URL ?= root:4521822123@tcp(127.0.0.1:3306)/rudy_gc
GOCTL  ?= goctl
STYLE  ?= go_zero
MOVIE_MODEL_DIR  ?= data/modelx/moviex
SPIDER_MODEL_DIR ?= data/modelx/spiderx
MOVIE_TABLES  ?= a*,am*,bm*,c*,e*,v*,g*
SPIDER_TABLES ?= d*

# ---- 将可能的相对路径归一化为绝对路径（相对项目根）----
ABS_MOVIE_MODEL_DIR  := $(if $(filter /%,$(MOVIE_MODEL_DIR)),  $(MOVIE_MODEL_DIR),  $(ROOT_DIR)/$(MOVIE_MODEL_DIR))
ABS_SPIDER_MODEL_DIR := $(if $(filter /%,$(SPIDER_MODEL_DIR)), $(SPIDER_MODEL_DIR), $(ROOT_DIR)/$(SPIDER_MODEL_DIR))

# =========================================
# 🎬 生成 movie 模型（带缓存）
# =========================================
gen-model-movie:
	@echo "🎬 Generating movie models into: $(ABS_MOVIE_MODEL_DIR)"
	@$(GOCTL) model mysql datasource \
		-url='$(DB_URL)' \
		-table='$(MOVIE_TABLES)' \
		-dir='$(ABS_MOVIE_MODEL_DIR)' \
		--style='$(STYLE)' \
		-c
	@echo "✅ Movie models generated at: $(ABS_MOVIE_MODEL_DIR)"

# =========================================
# 🕷 生成 spider 模型（不带缓存）
# =========================================
gen-model-spider:
	@echo "🕷 Generating spider models into: $(ABS_SPIDER_MODEL_DIR)"
	@$(GOCTL) model mysql datasource \
		-url='$(DB_URL)' \
		-table='$(SPIDER_TABLES)' \
		-dir='$(ABS_SPIDER_MODEL_DIR)' \
		--style='$(STYLE)'
	@echo "✅ Spider models generated at: $(ABS_SPIDER_MODEL_DIR)"

# =========================================
# 📦 一次生成全部模型
# =========================================
gen-model: gen-model-movie gen-model-spider

# =========================================
# 🧹 清理生成目录
# =========================================
clean-model:
	@echo "🧹 Removing generated model dirs:"
	@echo "   - $(ABS_MOVIE_MODEL_DIR)"
	@echo "   - $(ABS_SPIDER_MODEL_DIR)"
	rm -rf "$(ABS_MOVIE_MODEL_DIR)" "$(ABS_SPIDER_MODEL_DIR)"
	@echo "✅ Clean complete."