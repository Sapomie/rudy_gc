ifndef PROJECT_BACKUP_MK
PROJECT_BACKUP_MK := 1

PROJECT_BACKUP_FILE := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_BACKUP_MAKE_DIR := $(dir $(PROJECT_BACKUP_FILE))

include $(PROJECT_BACKUP_MAKE_DIR)common.mk

.PHONY: backup-help db-backup backup

GC_LOCAL_DB_HOST ?= 127.0.0.1
GC_LOCAL_DB_PORT ?= 3306
GC_LOCAL_DB_USER ?= root
GC_LOCAL_DB_PASS ?= 4521822123
GC_LOCAL_DB_NAME ?= rudy_gc
GC_BACKUP_DIR ?= /Volumes/T7/va/backup/sql/rudy_gc
GC_SQL_TS := $(shell TZ=Asia/Shanghai date '+%Y%m%d_%H%M%S')
GC_BACKUP_FILE ?= $(GC_BACKUP_DIR)/rudy_gc_$(GC_SQL_TS).sql

backup-help:
	@echo "Backup targets:"
	@echo "  make db-backup  Backup local rudy_gc MySQL database"
	@echo "  make backup     Alias of db-backup"
	@echo "Backup variables:"
	@echo "  GC_LOCAL_DB_HOST=$(GC_LOCAL_DB_HOST)"
	@echo "  GC_LOCAL_DB_PORT=$(GC_LOCAL_DB_PORT)"
	@echo "  GC_LOCAL_DB_USER=$(GC_LOCAL_DB_USER)"
	@echo "  GC_LOCAL_DB_NAME=$(GC_LOCAL_DB_NAME)"
	@echo "  GC_BACKUP_DIR=$(GC_BACKUP_DIR)"
	@echo "  GC_BACKUP_FILE=$(GC_BACKUP_FILE)"

db-backup:
	@echo "Backing up local rudy_gc database: $(GC_LOCAL_DB_NAME)"
	@echo "Output file: $(GC_BACKUP_FILE)"
	@mkdir -p "$(GC_BACKUP_DIR)"
	@tmp_file="$(GC_BACKUP_FILE).tmp"; \
	cleanup() { rm -f "$$tmp_file"; }; \
	trap cleanup EXIT; \
	MYSQL_PWD="$(GC_LOCAL_DB_PASS)" mysqldump \
		--host="$(GC_LOCAL_DB_HOST)" \
		--port="$(GC_LOCAL_DB_PORT)" \
		--user="$(GC_LOCAL_DB_USER)" \
		--single-transaction \
		--quick \
		--routines \
		--triggers \
		--events \
		--default-character-set=utf8mb4 \
		--set-gtid-purged=OFF \
		"$(GC_LOCAL_DB_NAME)" > "$$tmp_file"; \
	mv "$$tmp_file" "$(GC_BACKUP_FILE)"
	@echo "backup_file=$(GC_BACKUP_FILE)"

backup: db-backup

endif
