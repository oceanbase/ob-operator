#!/bin/sh
# Read-only viewer; installation remains a native Helm command.
set -eu
for interpreter in python3 python3.14 python3.13 python3.12 python3.11 python3.10 python3.9 python3.8; do
  if command -v "$interpreter" >/dev/null 2>&1 &&
     "$interpreter" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 8) else 1)' >/dev/null 2>&1; then
    exec "$interpreter" "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/progress.py" "$@"
  fi
done
echo 'Python 3.8+ is required. Install a supported Python 3 interpreter and retry.' >&2
exit 2
