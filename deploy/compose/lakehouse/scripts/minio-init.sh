#!/usr/bin/env sh
set -eu

access_id="$(cat /run/secrets/minio_access_id)"
access_key="$(cat /run/secrets/minio_access_key)"
until mc alias set lakehouse http://minio:9000 "${access_id}" "${access_key}" >/dev/null 2>&1; do
  sleep 2
done
mc mb --ignore-existing "lakehouse/${LOGSERVICE_BUCKET}"
mc mb --ignore-existing "lakehouse/${SHARED_STORAGE_BUCKET}"
