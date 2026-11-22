#!/bin/sh
set -eu
payload=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
exec python3 "$payload/bootstrap.py" --payload "$payload" "$@"
