#!/usr/bin/env bash
set -euo pipefail

socket=${OBSHELL_SOCKET:-/home/admin/oceanbase/run/obshell.sock}
prometheus_url=${PROMETHEUS_INTERNAL_URL:-http://prometheus:9090/}
alertmanager_url=${ALERTMANAGER_INTERNAL_URL:-http://alertmanager:9093}
timeout_seconds=${MONITORING_REGISTER_TIMEOUT_SECONDS:-600}
[[ "${timeout_seconds}" =~ ^[1-9][0-9]*$ ]] || {
  echo "MONITORING_REGISTER_TIMEOUT_SECONDS must be a positive integer" >&2
  exit 1
}
deadline=$((SECONDS + timeout_seconds))

wait_ready() {
  local name=$1
  local url=$2
  until curl -fsS --max-time 3 "${url%/}/-/ready" >/dev/null 2>&1; do
    if (( SECONDS >= deadline )); then
      echo "Timed out waiting for ${name}; automatic obshell registration was skipped" >&2
      return 1
    fi
    sleep 2
  done
}

register() {
  local name=$1
  local url=$2
  local response
  response=$(curl -fsS --max-time 10 --unix-socket "${socket}" \
    -X PUT "http://localhost/api/v1/system/external/${name}" \
    -H 'Content-Type: application/json' \
    --data-binary "{\"address\":\"${url}\",\"auth\":{\"username\":\"\",\"password\":\"\"}}")
  if [[ "${response}" != *'"successful":true'* ]]; then
    echo "obshell rejected ${name} registration: ${response}" >&2
    return 1
  fi
  echo "obshell ${name} registered: ${url}"
}

while [[ ! -S "${socket}" ]]; do
  if (( SECONDS >= deadline )); then
    echo "Timed out waiting for obshell socket; automatic monitoring registration was skipped" >&2
    exit 1
  fi
  sleep 1
done

wait_ready prometheus "${prometheus_url}"
wait_ready alertmanager "${alertmanager_url}"
register prometheus "${prometheus_url}"
register alertmanager "${alertmanager_url}"
