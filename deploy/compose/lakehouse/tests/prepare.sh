#!/usr/bin/env bash
set -euo pipefail
bundle_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture_dir="$(mktemp -d "${TMPDIR:-/tmp}/lakehouse-prepare-test.XXXXXX")"
trap 'rm -rf -- "${fixture_dir:?}"' EXIT
mkdir -p "${fixture_dir}/scripts"
cp "${bundle_dir}/scripts/prepare.sh" "${fixture_dir}/scripts/prepare.sh"
cp "${bundle_dir}/.env.example" "${fixture_dir}/.env.example"
"${fixture_dir}/scripts/prepare.sh"
before="$(cksum "${fixture_dir}/.env" "${fixture_dir}/secrets/"*)"
"${fixture_dir}/scripts/prepare.sh"
if [[ "${before}" != "$(cksum "${fixture_dir}/.env" "${fixture_dir}/secrets/"*)" ]]; then
  echo "Repeated prepare changed existing configuration or credentials" >&2
  exit 1
fi
for credential in minio_access_id minio_access_key ob_root_password; do
  file="${fixture_dir}/secrets/${credential}"
  [[ -s "${file}" ]]
  if [[ "$(uname -s)" == Darwin ]]; then
    mode="$(stat -f '%Lp' "${file}")"
  else
    mode="$(stat -c '%a' "${file}")"
  fi
  [[ "${mode}" == 600 ]]
done
echo "PASS: credential creation, owner-only permissions and repeatability"
