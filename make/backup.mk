# make/backup.mk
ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

BACKUP_SCRIPT ?= $(ROOT_DIR)/deploy/scripts/backup_db.sh
BACKUP_DIR    ?= /Volumes/T7/va/backup/sql/rudy-gc

.PHONY: db-backup
db-backup:
	@echo "🗄️  Backing up database: $(DB_URL)"
	@echo "📁 Output dir: $(BACKUP_DIR)"
	@DB_URL='$(DB_URL)' BACKUP_DIR='$(BACKUP_DIR)' bash '$(BACKUP_SCRIPT)'
