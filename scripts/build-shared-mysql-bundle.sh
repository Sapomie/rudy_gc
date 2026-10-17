#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <release-tag> <output-dir>" >&2
  exit 2
fi

release_tag="$1"
output_dir="$2"
workspace_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target_image="localhost/rudy_infra/mysql:${release_tag}"
builder_args=()
if [[ -n "${BUILDX_BUILDER:-}" ]]; then
  builder_args+=(--builder "$BUILDX_BUILDER")
fi

if [[ ! "$release_tag" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]*$ ]]; then
  echo "invalid release tag: $release_tag" >&2
  exit 2
fi

if [[ -d "$output_dir" ]] && [[ -n "$(find "$output_dir" -mindepth 1 -maxdepth 1 -print -quit)" ]]; then
  echo "output directory is not empty: $output_dir" >&2
  exit 1
fi

mkdir -p "$output_dir"

docker buildx build "${builder_args[@]}" --platform linux/amd64 --load \
  -f "$workspace_root/deploy/shared/mysql/image.Dockerfile" \
  -t "$target_image" "$workspace_root"

architecture="$(docker image inspect --format '{{.Architecture}}' "$target_image")"
if [[ "$architecture" != "amd64" ]]; then
  echo "unexpected image architecture: $architecture" >&2
  exit 1
fi

docker save "$target_image" | gzip >"$output_dir/mysql-images.tar.gz"
cp "$workspace_root/deploy/shared/mysql/compose.yaml" "$output_dir/mysql.compose.yaml"
cp "$workspace_root/scripts/install-shared-mysql.sh" "$output_dir/install-shared-mysql.sh"
chmod 700 "$output_dir/install-shared-mysql.sh"

printf 'MYSQL_IMAGE=%s\n' "$target_image" >"$output_dir/mysql-release.env"
docker image inspect --format '{{.RepoTags}} {{.Id}} {{.Os}}/{{.Architecture}}' "$target_image" >"$output_dir/IMAGE_MANIFEST"

(
  cd "$output_dir"
  shasum -a 256 IMAGE_MANIFEST install-shared-mysql.sh mysql-images.tar.gz mysql-release.env mysql.compose.yaml >SHA256SUMS
)

echo "shared MySQL offline bundle: $output_dir"
