#!/usr/bin/env bash
set -euo pipefail

name="${1:?Usage: link-local-plugin.sh <plugin-name> <repo-path>}"
path="${2:?Usage: link-local-plugin.sh <plugin-name> <repo-path>}"

root="$(cd "$(dirname "$0")/.." && pwd)"
target="${root}/.local-plugins/${name}"

if [ ! -d "$path" ]; then
  echo "error: repo path not found: $path" >&2
  exit 1
fi

mkdir -p "${root}/.local-plugins"
ln -sfn "$(cd "$path" && pwd)" "$target"

echo "Linked ${target} -> $(readlink "$target")"
echo "Restart backend: docker compose up -d --build krew-backend"
