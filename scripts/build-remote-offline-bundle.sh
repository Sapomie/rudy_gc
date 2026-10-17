#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <release-tag> <output-dir>" >&2
  exit 2
fi

release_tag="$1"
output_dir="$2"
workspace_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
api_image="localhost/rudy_gc/remote-api:${release_tag}"
web_image="localhost/rudy_gc/remote-web:${release_tag}"
solver_image="localhost/rudy_gc/remote-solver:${release_tag}"
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
  -f "$workspace_root/remote/deploy/docker/api.Dockerfile" \
  -t "$api_image" "$workspace_root"
docker buildx build "${builder_args[@]}" --platform linux/amd64 --load \
  -f "$workspace_root/remote/deploy/docker/web.Dockerfile" \
  -t "$web_image" "$workspace_root"
docker buildx build "${builder_args[@]}" --platform linux/amd64 --load \
  -f "$workspace_root/remote/deploy/docker/solver.Dockerfile" \
  -t "$solver_image" "$workspace_root"

for image in "$api_image" "$web_image" "$solver_image"; do
  architecture="$(docker image inspect --format '{{.Architecture}}' "$image")"
  if [[ "$architecture" != "amd64" ]]; then
    echo "unexpected image architecture for $image: $architecture" >&2
    exit 1
  fi
done

docker save "$api_image" "$web_image" "$solver_image" | gzip >"$output_dir/remote-images.tar.gz"
cp "$workspace_root/remote/deploy/docker/compose.yaml" "$output_dir/compose.yaml"
cp "$workspace_root/remote/deploy/docker/nginx.conf" "$output_dir/nginx.conf"
cp "$workspace_root/scripts/install-remote-offline.sh" "$output_dir/install-remote-offline.sh"
chmod 700 "$output_dir/install-remote-offline.sh"

{
  printf 'REMOTE_API_IMAGE=%s\n' "$api_image"
  printf 'REMOTE_WEB_IMAGE=%s\n' "$web_image"
  printf 'REMOTE_SOLVER_IMAGE=%s\n' "$solver_image"
} >"$output_dir/remote-release.env"

docker image inspect --format '{{.RepoTags}} {{.Id}} {{.Os}}/{{.Architecture}}' \
  "$api_image" "$web_image" "$solver_image" >"$output_dir/IMAGE_MANIFEST"

(
  cd "$output_dir"
  shasum -a 256 IMAGE_MANIFEST compose.yaml install-remote-offline.sh nginx.conf remote-images.tar.gz remote-release.env >SHA256SUMS
)

echo "Remote offline bundle: $output_dir"
