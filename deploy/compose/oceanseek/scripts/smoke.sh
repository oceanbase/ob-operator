#!/usr/bin/env bash
set -euo pipefail
base_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${base_dir}"
docker compose -f compose.yaml run --rm --no-deps observer-bootstrap --mode "${1:-smoke-read}"
