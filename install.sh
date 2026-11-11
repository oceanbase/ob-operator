#!/usr/bin/env bash
# Complete single-worker or three-worker lakehouse install, including Dashboard.
set -euo pipefail
export LC_ALL=C
bundle_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
installer="$bundle_dir/bin/lakehouse-installer"
if [[ ! -f "$bundle_dir/SHA256SUMS" ]]; then
  echo 'Package checksum manifest is missing. Extract the complete lakehouse tar again.' >&2
  exit 2
fi
if ! command -v sha256sum >/dev/null 2>&1; then
  echo 'Package verification requires sha256sum (Linux coreutils).' >&2
  exit 2
fi
verification_started=$SECONDS
echo '[verify] Verifying package integrity...'
if ! (cd -- "$bundle_dir" && sha256sum --check --quiet SHA256SUMS); then
  echo 'Package verification failed. Extract the complete tar again; put custom values in a separate file and pass -f.' >&2
  exit 2
fi
if [[ ! -x "$installer" ]]; then
  echo 'Package installer is missing or not executable. Extract the complete Linux amd64 lakehouse tar again.' >&2
  exit 2
fi
echo "[done] Package integrity verified ($((SECONDS - verification_started))s elapsed)"
if [[ -d "$bundle_dir/tools/bin" ]]; then
  export PATH="$bundle_dir/tools/bin:$PATH"
fi
exec "$installer" --bundle "$bundle_dir" "$@"
