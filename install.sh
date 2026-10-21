#!/usr/bin/env bash
# Complete single-worker lakehouse install, including Dashboard.
set -euo pipefail
bundle_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
installer="$bundle_dir/bin/lakehouse-installer"
if [[ ! -x "$installer" ]]; then
  echo 'This source checkout has no installer binary. Use the packaged Linux amd64 bundle, or build cmd/lakehouse-installer first.' >&2
  exit 2
fi
exec "$installer" --bundle "$bundle_dir" "$@"
