#!/usr/bin/env bash
set -euo pipefail

base_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
"${base_dir}/scripts/prepare.sh"
cd "${base_dir}"
set -a
# shellcheck disable=SC1091
source .env
set +a
docker compose up -d --build

bootstrap_id="$(docker compose ps -a -q observer-bootstrap)"
if [[ -z "${bootstrap_id}" ]]; then
  echo "observer-bootstrap container was not created" >&2
  exit 1
fi
while [[ "$(docker inspect -f '{{.State.Running}}' "${bootstrap_id}")" == "true" ]]; do
  sleep 2
done
exit_code="$(docker inspect -f '{{.State.ExitCode}}' "${bootstrap_id}")"
docker compose ps -a
if [[ "${exit_code}" != "0" ]]; then
  docker compose logs observer-bootstrap
  exit "${exit_code}"
fi
echo "Lakehouse is ready. Connect to 127.0.0.1:${OBSERVER_SQL_HOST_PORT:-32881} as root with an empty password."
