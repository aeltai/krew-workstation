#!/usr/bin/env bash
set -euo pipefail
exec "$(dirname "$0")/build-local-cli-for-docker.sh" rk9s "${1:-/Users/aeltai/k9s/k9s}"
