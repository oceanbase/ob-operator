#!/usr/bin/env bash
set -euo pipefail

base_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "${base_dir}/secrets"

if [[ ! -f "${base_dir}/.env" ]]; then
  cp "${base_dir}/.env.example" "${base_dir}/.env"
fi

generate_secret() {
  local path="$1"
  local bytes="$2"
  if [[ ! -s "${path}" ]]; then
    umask 077
    if command -v openssl >/dev/null 2>&1; then
      openssl rand -hex "${bytes}" >"${path}"
    else
      od -An -N "${bytes}" -tx1 /dev/urandom | tr -d ' \n' >"${path}"
      printf '\n' >>"${path}"
    fi
  fi
}

generate_secret "${base_dir}/secrets/minio_access_id" 10
generate_secret "${base_dir}/secrets/minio_access_key" 24
chmod 600 "${base_dir}/secrets/minio_access_id" "${base_dir}/secrets/minio_access_key"

echo "Lakehouse Compose environment is prepared in ${base_dir}"
