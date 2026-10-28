#!/usr/bin/env bash
# Maintainer test: restarts/recreates only an explicitly named QA project.
set -euo pipefail
base_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${base_dir}"
if [[ "${COMPOSE_PROJECT_NAME:-}" != ob-lakehouse-qa-* ]]; then
  echo "Set COMPOSE_PROJECT_NAME=ob-lakehouse-qa-<identifier> and use unused ports/subnet." >&2
  exit 1
fi

timeout_seconds="${STARTUP_TIMEOUT_SECONDS:-2400}"
if [[ ! "${timeout_seconds}" =~ ^[1-9][0-9]*$ ]]; then
  echo "STARTUP_TIMEOUT_SECONDS must be a positive integer" >&2
  exit 1
fi

start_and_check() {
  docker compose up -d
  local container_id deadline status
  container_id="$(docker compose ps -a -q observer-bootstrap)"
  [[ -n "${container_id}" ]] || { echo "Observer bootstrap container was not created" >&2; return 1; }
  deadline=$((SECONDS + timeout_seconds))
  while true; do
    status="$(docker inspect -f '{{.State.Status}}' "${container_id}")"
    [[ "${status}" == "exited" ]] && break
    if (( SECONDS >= deadline )); then
      echo "Initialization timed out; inspect docker compose ps -a and logs." >&2
      return 1
    fi
    sleep 2
  done
  if [[ "$(docker inspect -f '{{.State.ExitCode}}' "${container_id}")" != "0" ]]; then
    docker compose logs observer-bootstrap
    return 1
  fi
  ./scripts/smoke.sh check
}

docker compose config --quiet
start_and_check
password_before="$(docker compose run -T --rm --no-deps observer-bootstrap --mode show-password)"
./scripts/smoke.sh smoke-write
docker compose run --rm --no-deps observer-bootstrap --mode bootstrap
if docker compose run --rm --no-deps -e ROOT_PASSWORD_FILE=/run/secrets/minio_access_key observer-bootstrap --mode check; then
  echo "Wrong database password was accepted" >&2
  exit 1
fi
docker compose restart logservice-1 logservice-2 logservice-3 observer
./scripts/smoke.sh smoke-read
docker compose down
start_and_check
./scripts/smoke.sh smoke-read
password_after="$(docker compose run -T --rm --no-deps observer-bootstrap --mode show-password)"
if [[ "${password_before}" != "${password_after}" ]]; then
  echo "Credential changed across repeated starts" >&2
  exit 1
fi
unset password_before password_after
echo "PASS: authenticated startup, repeated bootstrap, rejected wrong password, restart and down/up persistence"
