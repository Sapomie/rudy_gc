#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 <bundle-dir> <mysql-env-file> <mysql-data-dir>" >&2
  exit 2
fi

bundle_dir="$(cd "$1" && pwd)"
MYSQL_ENV_FILE="$2"
MYSQL_DATA_DIR="$3"

cd "$bundle_dir"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum -c SHA256SUMS
else
  shasum -a 256 -c SHA256SUMS
fi

set -a
source "$bundle_dir/mysql-release.env"
source "$MYSQL_ENV_FILE"
set +a

if [[ -z "${MYSQL_ROOT_PASSWORD:-}" ]]; then
  echo "MYSQL_ROOT_PASSWORD is required" >&2
  exit 1
fi

gunzip -c "$bundle_dir/mysql-images.tar.gz" | docker load

architecture="$(docker image inspect --format '{{.Architecture}}' "$MYSQL_IMAGE")"
if [[ "$architecture" != "amd64" ]]; then
  echo "unexpected image architecture: $architecture" >&2
  exit 1
fi

if ! docker network inspect rudy_shared_infra >/dev/null 2>&1; then
  docker network create rudy_shared_infra >/dev/null
fi

install -d -m 700 "$MYSQL_DATA_DIR"
export MYSQL_ENV_FILE MYSQL_DATA_DIR MYSQL_IMAGE

docker compose -p rudy-shared-mysql --env-file "$bundle_dir/mysql-release.env" -f "$bundle_dir/mysql.compose.yaml" config >/dev/null
docker compose -p rudy-shared-mysql --env-file "$bundle_dir/mysql-release.env" -f "$bundle_dir/mysql.compose.yaml" up -d --pull never --no-build

until docker exec rudy_mysql_shared mysqladmin ping -h 127.0.0.1 -uroot -p"$MYSQL_ROOT_PASSWORD" --silent >/dev/null 2>&1; do
  sleep 2
done

docker ps --filter name='^/rudy_mysql_shared$' --format '{{.Names}} {{.Image}} {{.Status}}'
