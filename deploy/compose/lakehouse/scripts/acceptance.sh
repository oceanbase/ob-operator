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

wait_healthy() {
  local service="$1" container_id deadline health
  container_id="$(docker compose ps -q "${service}")"
  [[ -n "${container_id}" ]] || { echo "${service} container was not created" >&2; return 1; }
  deadline=$((SECONDS + timeout_seconds))
  while true; do
    health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${container_id}")"
    [[ "${health}" == "healthy" ]] && return 0
    if [[ "${health}" == "unhealthy" ]] || (( SECONDS >= deadline )); then
      echo "${service} did not become healthy (status=${health})" >&2
      docker compose logs --tail=100 "${service}" >&2
      return 1
    fi
    sleep 2
  done
}

check_monitoring() {
  wait_healthy obagent
  wait_healthy prometheus
  wait_healthy alertmanager
  docker compose exec -T observer bash -c 'exec 3<>/dev/tcp/127.0.0.1/2886'
  docker compose exec -T obagent curl -fsS http://127.0.0.1:8088/metrics/stat >/dev/null
  docker compose exec -T obagent curl -fsS http://127.0.0.1:8088/metrics/ob/basic >/dev/null
  docker compose exec -T alertmanager wget -qO- http://127.0.0.1:9093/-/ready >/dev/null
  local endpoint expected response
  for endpoint in prometheus alertmanager; do
    if [[ "${endpoint}" == prometheus ]]; then
      expected='http://prometheus:9090/'
    else
      expected='http://alertmanager:9093'
    fi
    response="$(docker compose exec -T observer curl -fsS \
      --unix-socket /home/admin/oceanbase/run/obshell.sock \
      "http://localhost/api/v1/system/external/${endpoint}")"
    grep -Fq '"successful":true' <<<"${response}" || {
      echo "obshell ${endpoint} registration was not successful: ${response}" >&2
      return 1
    }
    grep -Fq "\"data\":\"${expected}\"" <<<"${response}" || {
      echo "obshell ${endpoint} address is not ${expected}: ${response}" >&2
      return 1
    }
  done
  local targets up_count deadline
  deadline=$((SECONDS + timeout_seconds))
  while true; do
    targets="$(docker compose exec -T prometheus wget -qO- http://127.0.0.1:9090/api/v1/targets)"
    up_count="$( { grep -o '"health":"up"' <<<"${targets}" || true; } | wc -l | tr -d ' ')"
    (( up_count >= 2 )) && break
    if (( SECONDS >= deadline )); then
      echo "Prometheus has fewer than two UP OBAgent scrape targets" >&2
      echo "${targets}" >&2
      return 1
    fi
    sleep 2
  done
}

wait_obshell_active() {
  local deadline output status
  deadline=$((SECONDS + timeout_seconds))
  while true; do
    output="$(docker compose exec -T observer /home/admin/oceanbase/bin/obshell cluster show 2>&1 || true)"
    status="$(docker compose exec -T observer curl -fsS \
      --unix-socket /home/admin/oceanbase/run/obshell.sock \
      http://localhost/api/v1/obcluster/info 2>/dev/null || true)"
    if grep -Eq '[[:space:]]ACTIVE[[:space:]]' <<<"${output}" && \
       grep -Eq '"status"[[:space:]]*:[[:space:]]*"AVAILABLE"' <<<"${status}"; then
      return 0
    fi
    if (( SECONDS >= deadline )); then
      echo "obshell did not take over the initialized cluster" >&2
      echo "${output}" >&2
      return 1
    fi
    sleep 2
  done
}

wait_smoke() {
  local mode="$1" deadline output
  deadline=$((SECONDS + timeout_seconds))
  while true; do
    if output="$(./scripts/smoke.sh "${mode}" 2>&1)"; then
      printf '%s\n' "${output}"
      return 0
    fi
    if (( SECONDS >= deadline )); then
      echo "${mode} did not succeed before the acceptance deadline" >&2
      echo "${output}" >&2
      return 1
    fi
    sleep 5
  done
}

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
  wait_obshell_active
  check_monitoring
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
wait_healthy logservice-1
wait_healthy logservice-2
wait_healthy logservice-3
wait_healthy observer
wait_healthy obagent
wait_obshell_active
wait_smoke smoke-read
docker compose down
start_and_check
wait_smoke smoke-read
password_after="$(docker compose run -T --rm --no-deps observer-bootstrap --mode show-password)"
if [[ "${password_before}" != "${password_after}" ]]; then
  echo "Credential changed across repeated starts" >&2
  exit 1
fi
unset password_before password_after
echo "PASS: authenticated startup, obshell, OBAgent metrics, Prometheus targets, Alertmanager readiness, repeated bootstrap, rejected wrong password, restart and down/up persistence"
