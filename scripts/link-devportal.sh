#!/usr/bin/env bash
# Link Developer Portal extension into this repo for local dev (both show in one Rancher UI).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEVPORTAL="${DEVPORTAL_ROOT:-$ROOT/../rancher-devportal/pkg/devportal}"
TARGET="$ROOT/pkg/devportal"

if [ ! -d "$DEVPORTAL" ]; then
  echo "Developer Portal not found at $DEVPORTAL"
  echo "Clone https://github.com/aeltai/rancher-devportal or set DEVPORTAL_ROOT"
  exit 1
fi

rm -f "$TARGET"
ln -sfn "$(cd "$DEVPORTAL" && pwd)" "$TARGET"
echo "Linked $TARGET -> $(readlink "$TARGET")"
echo "Run 'yarn dev' once — Krew + Developer Portal both appear in the sidebar."
