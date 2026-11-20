#!/usr/bin/env sh
set -eu

access_id="$(cat /run/secrets/minio_access_id)"
access_key="$(cat /run/secrets/minio_access_key)"
attempt=0
until mc alias set lakehouse http://minio:9000 "${access_id}" "${access_key}" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "${attempt}" -ge 150 ]; then
    echo "MinIO initialization failed after 150 attempts" >&2
    exit 1
  fi
  sleep 2
done
mc mb --ignore-existing "lakehouse/${LOGSERVICE_BUCKET}"
mc mb --ignore-existing "lakehouse/${SHARED_STORAGE_BUCKET}"
