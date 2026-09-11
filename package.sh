#!/usr/bin/env bash
# Native build by default. Cross builds require GOOS, GOARCH and a matching CC.
set -euo pipefail
cd "$(dirname "$0")"
exec ./scripts/build-release.sh "$@"
