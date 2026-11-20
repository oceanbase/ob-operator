#!/usr/bin/env sh
set -eu

MINIO_ROOT_USER="$(cat /run/secrets/minio_access_id)"
MINIO_ROOT_PASSWORD="$(cat /run/secrets/minio_access_key)"
export MINIO_ROOT_USER MINIO_ROOT_PASSWORD
exec minio server /data --console-address :9001
