#!/usr/bin/env bash

set -euo pipefail

DEFAULT_DB_URL='root:4521822123@tcp(127.0.0.1:3306)/rudy_gc'
DEFAULT_BACKUP_DIR='/Volumes/T7/va/backup/sql/rudy-gc/'

DB_URL="${DB_URL:-$DEFAULT_DB_URL}"
BACKUP_DIR="${BACKUP_DIR:-$DEFAULT_BACKUP_DIR}"

trim_trailing_slash() {
    local value="$1"
    value="${value%/}"
    printf '%s' "$value"
}

parse_db_url() {
    local url="$1"

    DB_USER_PARSED="$(printf '%s' "$url" | sed -E 's#^([^:]+):.*#\1#')"
    DB_PASSWORD_PARSED="$(printf '%s' "$url" | sed -E 's#^[^:]+:([^@]+)@tcp\(.*#\1#')"
    DB_HOST_PARSED="$(printf '%s' "$url" | sed -E 's#^.*@tcp\(([^:]+):[0-9]+\)/.*#\1#')"
    DB_PORT_PARSED="$(printf '%s' "$url" | sed -E 's#^.*@tcp\([^:]+:([0-9]+)\)/.*#\1#')"
    DB_NAME_PARSED="$(printf '%s' "$url" | sed -E 's#^.*/([^?]+)(\?.*)?$#\1#')"
}

parse_db_url "$DB_URL"

DB_HOST="${DB_HOST:-$DB_HOST_PARSED}"
DB_PORT="${DB_PORT:-$DB_PORT_PARSED}"
DB_USER="${DB_USER:-$DB_USER_PARSED}"
DB_PASSWORD="${DB_PASSWORD:-$DB_PASSWORD_PARSED}"
DB_NAME="${DB_NAME:-$DB_NAME_PARSED}"

BACKUP_DIR="$(trim_trailing_slash "$BACKUP_DIR")"
TIMESTAMP="$(date '+%Y-%m-%d_%H%M%S')"
OUTPUT_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql"
TMP_FILE="${OUTPUT_FILE}.tmp"

mkdir -p "$BACKUP_DIR"

cleanup() {
    if [ -f "$TMP_FILE" ]; then
        rm -f "$TMP_FILE"
    fi
}

trap cleanup EXIT

export MYSQL_PWD="$DB_PASSWORD"

mysqldump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --user="$DB_USER" \
    --single-transaction \
    --quick \
    --routines \
    --triggers \
    --events \
    --default-character-set=utf8mb4 \
    --set-gtid-purged=OFF \
    "$DB_NAME" > "$TMP_FILE"

mv "$TMP_FILE" "$OUTPUT_FILE"

echo "backup_file=$OUTPUT_FILE"
