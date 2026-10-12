#!/usr/bin/env bash
set -euo pipefail

base_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
"${base_dir}/scripts/prepare.sh"
cd "${base_dir}"
set -a
# shellcheck disable=SC1091
source .env
set +a
timeout_seconds="${STARTUP_TIMEOUT_SECONDS:-2400}"
if [[ ! "${timeout_seconds}" =~ ^[1-9][0-9]*$ ]]; then
  echo "STARTUP_TIMEOUT_SECONDS must be a positive integer" >&2
  exit 1
fi
if [[ "${BUILD_BOOTSTRAP:-1}" != "0" && "${BUILD_BOOTSTRAP:-1}" != "1" ]]; then
  echo "BUILD_BOOTSTRAP must be 0 or 1" >&2
  exit 1
fi
compose_args=(-f compose.yaml)
if [[ "${BUILD_BOOTSTRAP:-1}" == "1" ]]; then
  compose_args+=(-f compose.build.yaml)
fi
docker compose "${compose_args[@]}" config --quiet
if [[ "${BUILD_BOOTSTRAP:-1}" == "1" ]]; then
  docker compose "${compose_args[@]}" build observer-bootstrap
fi
docker compose "${compose_args[@]}" up -d

bootstrap_id="$(docker compose "${compose_args[@]}" ps -a -q observer-bootstrap)"
if [[ -z "${bootstrap_id}" ]]; then
  echo "observer-bootstrap container was not created" >&2
  exit 1
fi
deadline=$((SECONDS + timeout_seconds))
while [[ "$(docker inspect -f '{{.State.Status}}' "${bootstrap_id}")" != "exited" ]]; do
  if (( SECONDS >= deadline )); then
    echo "Observer initialization timed out; inspect Compose status and logs" >&2
    docker compose "${compose_args[@]}" ps -a
    exit 1
  fi
  sleep 2
done
exit_code="$(docker inspect -f '{{.State.ExitCode}}' "${bootstrap_id}")"
docker compose "${compose_args[@]}" ps -a
if [[ "${exit_code}" != "0" ]]; then
  docker compose "${compose_args[@]}" logs observer-bootstrap
  exit "${exit_code}"
fi
docker compose "${compose_args[@]}" run --rm --no-deps observer-bootstrap --mode check
echo "Lakehouse is ready at ${HOST_BIND_ADDRESS:-127.0.0.1}:${OBSERVER_SQL_HOST_PORT:-32881}. Root password: secrets/ob_root_password (do not share this file)."
