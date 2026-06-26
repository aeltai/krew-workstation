#!/usr/bin/env bash
set -euo pipefail

name="${1:?Usage: build-local-cli-for-docker.sh <binary-name> [source-dir]}"
src="${2:-}"

case "$name" in
  rk9s)
    src="${src:-/Users/aeltai/k9s/k9s}"
    ;;
  rancher-polymorph)
    src="${src:-/Users/aeltai/rancher-migrate}"
    ;;
  *)
    src="${src:-}"
    ;;
esac

if [ -z "$src" ] || [ ! -f "${src}/go.mod" ]; then
  echo "error: source dir with go.mod required for ${name}" >&2
  echo "usage: $0 ${name} /path/to/source" >&2
  exit 1
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
plugin_dir="${root}/.local-plugins/${name}"
out_dir="${plugin_dir}/bin"
out="${out_dir}/${name}"

if [ -L "$plugin_dir" ]; then
  rm "$plugin_dir"
fi
mkdir -p "$out_dir"

echo "Building ${name} for linux/amd64 from ${src}..."
(
  cd "$src"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$out" .
)

echo "Built ${out} ($(file -b "$out"))"
echo "Restart backend: docker compose restart krew-backend"
