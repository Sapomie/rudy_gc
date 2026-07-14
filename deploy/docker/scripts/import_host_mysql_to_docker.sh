#!/usr/bin/env bash

set -euo pipefail

TARGET_CONTAINER="${TARGET_CONTAINER:-rudy_infra_mysql_local}"

SOURCE_HOST="${SOURCE_HOST:-host.docker.internal}"
SOURCE_PORT="${SOURCE_PORT:-3306}"
SOURCE_USER="${SOURCE_USER:-root}"
SOURCE_PASSWORD="${SOURCE_PASSWORD:-4521822123}"
SOURCE_DB="${SOURCE_DB:-rudy_gc}"

TARGET_HOST="${TARGET_HOST:-127.0.0.1}"
TARGET_PORT="${TARGET_PORT:-3306}"
TARGET_USER="${TARGET_USER:-root}"
TARGET_PASSWORD="${TARGET_PASSWORD:-4521822123}"
TARGET_DB="${TARGET_DB:-rudy_gc}"

sql_ident() {
    printf '`%s`' "$(printf '%s' "$1" | sed 's/`/``/g')"
}

sql_string() {
    printf "'%s'" "$(printf '%s' "$1" | sed "s/'/''/g")"
}

require_container() {
    if ! docker ps --format '{{.Names}}' | grep -Fxq "$TARGET_CONTAINER"; then
        echo "target_container_missing=$TARGET_CONTAINER" >&2
        exit 1
    fi
}

require_healthy_container() {
    local health_status
    health_status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$TARGET_CONTAINER")"
    if [ "$health_status" != "healthy" ] && [ "$health_status" != "none" ]; then
        echo "target_container_health=$health_status" >&2
        exit 1
    fi
}

require_mysql_tools() {
    docker exec "$TARGET_CONTAINER" sh -c 'command -v mysql >/dev/null && command -v mysqladmin >/dev/null && command -v mysqldump >/dev/null'
}

require_source_mysql() {
    docker exec -e MYSQL_PWD="$SOURCE_PASSWORD" "$TARGET_CONTAINER" \
        mysqladmin ping \
        --host="$SOURCE_HOST" \
        --port="$SOURCE_PORT" \
        --user="$SOURCE_USER" \
        --protocol=tcp \
        --silent >/dev/null
}

require_target_mysql() {
    docker exec -e MYSQL_PWD="$TARGET_PASSWORD" "$TARGET_CONTAINER" \
        mysqladmin ping \
        --host="$TARGET_HOST" \
        --port="$TARGET_PORT" \
        --user="$TARGET_USER" \
        --protocol=tcp \
        --silent >/dev/null
}

create_target_database() {
    local create_sql
    create_sql="CREATE DATABASE IF NOT EXISTS $(sql_ident "$TARGET_DB") CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    docker exec -e MYSQL_PWD="$TARGET_PASSWORD" "$TARGET_CONTAINER" \
        mysql \
        --host="$TARGET_HOST" \
        --port="$TARGET_PORT" \
        --user="$TARGET_USER" \
        --protocol=tcp \
        --execute="$create_sql"
}

require_empty_target_database() {
    local count_sql table_count
    count_sql="SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $(sql_string "$TARGET_DB");"
    table_count="$(docker exec -e MYSQL_PWD="$TARGET_PASSWORD" "$TARGET_CONTAINER" \
        mysql \
        --host="$TARGET_HOST" \
        --port="$TARGET_PORT" \
        --user="$TARGET_USER" \
        --protocol=tcp \
        --batch \
        --skip-column-names \
        --execute="$count_sql")"
    if [ "${table_count:-0}" != "0" ]; then
        echo "target_database_has_tables=$table_count" >&2
        exit 1
    fi
}

import_database() {
    docker exec -e MYSQL_PWD="$SOURCE_PASSWORD" "$TARGET_CONTAINER" \
        mysqldump \
        --host="$SOURCE_HOST" \
        --port="$SOURCE_PORT" \
        --user="$SOURCE_USER" \
        --protocol=tcp \
        --single-transaction \
        --quick \
        --routines \
        --triggers \
        --events \
        --default-character-set=utf8mb4 \
        --set-gtid-purged=OFF \
        --skip-add-drop-table \
        --skip-add-drop-database \
        "$SOURCE_DB" \
    | docker exec -i -e MYSQL_PWD="$TARGET_PASSWORD" "$TARGET_CONTAINER" \
        mysql \
        --host="$TARGET_HOST" \
        --port="$TARGET_PORT" \
        --user="$TARGET_USER" \
        --protocol=tcp \
        --default-character-set=utf8mb4 \
        "$TARGET_DB"
}

require_container
require_healthy_container
require_mysql_tools
require_source_mysql
require_target_mysql
create_target_database
require_empty_target_database
import_database

echo "imported ${SOURCE_HOST}:${SOURCE_PORT}/${SOURCE_DB} -> ${TARGET_CONTAINER}:${TARGET_PORT}/${TARGET_DB}"
