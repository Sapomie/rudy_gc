#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <bundle-dir> <server-config-file>" >&2
  exit 2
fi

bundle_dir="$(cd "$1" && pwd)"
REMOTE_CONFIG_FILE="$2"
REMOTE_NGINX_CONFIG_FILE="$bundle_dir/nginx.conf"

cd "$bundle_dir"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum -c SHA256SUMS
else
  shasum -a 256 -c SHA256SUMS
fi

set -a
source "$bundle_dir/remote-release.env"
set +a

if [[ ! -s "$REMOTE_CONFIG_FILE" ]]; then
  echo "server config is missing or empty: $REMOTE_CONFIG_FILE" >&2
  exit 1
fi
if [[ ! -s "$REMOTE_NGINX_CONFIG_FILE" ]]; then
  echo "Nginx config is missing or empty: $REMOTE_NGINX_CONFIG_FILE" >&2
  exit 1
fi
if ! docker network inspect rudy_shared_infra >/dev/null 2>&1; then
  echo "shared Docker network is missing: rudy_shared_infra" >&2
  exit 1
fi
if [[ "$(docker inspect --format '{{.State.Running}}' rudy_mysql_shared 2>/dev/null || true)" != "true" ]]; then
  echo "shared MySQL container is not running" >&2
  exit 1
fi

gunzip -c "$bundle_dir/remote-images.tar.gz" | docker load

for image in "$REMOTE_API_IMAGE" "$REMOTE_WEB_IMAGE" "$REMOTE_SOLVER_IMAGE"; do
  architecture="$(docker image inspect --format '{{.Architecture}}' "$image")"
  if [[ "$architecture" != "amd64" ]]; then
    echo "unexpected image architecture for $image: $architecture" >&2
    exit 1
  fi
done

export REMOTE_API_IMAGE REMOTE_WEB_IMAGE REMOTE_SOLVER_IMAGE REMOTE_CONFIG_FILE REMOTE_NGINX_CONFIG_FILE
compose=(docker compose -p rudy-gc-remote --env-file "$bundle_dir/remote-release.env" -f "$bundle_dir/compose.yaml")

"${compose[@]}" config >/dev/null
"${compose[@]}" run --rm --no-deps --pull never --entrypoint /app/remote-migrate remote-api -f /app/config/server.yaml
"${compose[@]}" up -d --pull never --no-build remote-solver
until curl -fsS http://172.30.0.1:8191/ >/dev/null; do
  sleep 2
done
"${compose[@]}" up -d --pull never --no-build remote-api
until curl -fsS http://127.0.0.1:5031/readyz >/dev/null; do
  sleep 2
done
"${compose[@]}" up -d --pull never --no-build remote-web
"${compose[@]}" ps
