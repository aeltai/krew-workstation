#!/usr/bin/env bash
set -euo pipefail
exec "$(dirname "$0")/build-local-cli-for-docker.sh" rancher-polymorph "${1:-/Users/aeltai/rancher-migrate}"
