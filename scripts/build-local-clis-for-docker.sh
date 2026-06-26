#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
"${root}/scripts/build-local-cli-for-docker.sh" rk9s
"${root}/scripts/build-local-cli-for-docker.sh" rancher-polymorph
