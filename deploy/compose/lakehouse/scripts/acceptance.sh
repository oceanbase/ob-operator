#!/usr/bin/env bash
# This test restarts and recreates only an explicitly named QA Compose project.
set -euo pipefail
base_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${base_dir}"
set -a
# shellcheck disable=SC1091
source .env
set +a
if [[ "${COMPOSE_PROJECT_NAME:-}" != ob-lakehouse-qa-* ]]; then
  echo "Use a separate ob-lakehouse-qa-* project, subnet, ports and volumes for acceptance" >&2
  exit 1
fi

secret_before="$(cksum secrets/ob_root_password)"
./scripts/up.sh
./scripts/smoke.sh smoke-write
docker compose run --rm --no-deps observer-bootstrap --mode bootstrap
if docker compose run --rm --no-deps -e ROOT_PASSWORD_FILE=/run/secrets/minio_access_key observer-bootstrap --mode check; then
  echo "Wrong database password was accepted" >&2
  exit 1
fi
./scripts/smoke.sh check
docker compose restart logservice-1 logservice-2 logservice-3 observer
./scripts/smoke.sh smoke-read
docker compose down
./scripts/up.sh
./scripts/smoke.sh smoke-read
if [[ "${secret_before}" != "$(cksum secrets/ob_root_password)" ]]; then
  echo "Credential changed across repeated preparation" >&2
  exit 1
fi
echo "PASS: authenticated startup, repeated bootstrap, rejected wrong password, restart and down/up persistence"
