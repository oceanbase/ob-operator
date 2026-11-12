#!/bin/sh
set -eu

mode=bootstrap
next_is_mode=false
for arg in "$@"; do
  if [ "${next_is_mode}" = true ]; then
    mode="${arg}"
    next_is_mode=false
    continue
  fi
  case "${arg}" in
    --mode) next_is_mode=true ;;
    --mode=*) mode="${arg#--mode=}" ;;
  esac
done

/usr/local/bin/lakehouse-bootstrap "$@"

if [ "${mode}" = bootstrap ]; then
  marker=/run/secrets/observer_bootstrap_complete
  temporary="$(mktemp "${marker}.tmp.XXXXXX")"
  trap 'rm -f "${temporary}"' EXIT HUP INT TERM
  printf 'complete\n' >"${temporary}"
  chmod 0600 "${temporary}"
  mv -f "${temporary}" "${marker}"
  trap - EXIT HUP INT TERM

  host_ip="$(cat /run/secrets/docker_host_ip)"
  shell_quote() {
    printf "'"
    printf '%s' "$1" | sed "s/'/'\\\\''/g"
    printf "'"
  }
  sql_host="$(shell_quote "${host_ip}")"
  compose_project="$(shell_quote "${COMPOSE_PROJECT_NAME:-ob-lakehouse}")"
  url_host="${host_ip}"
  case "${url_host}" in *:*) url_host="[${url_host}]" ;; esac

  cat <<EOF

Lakehouse deployment completed.

OceanBase SQL:
  mysql -h ${sql_host} -P ${OBSERVER_SQL_HOST_PORT:-32881} -u 'root@sys' -p
  Enter the password retrieved below when prompted.

obshell Dashboard:
  http://${url_host}:${OBSHELL_HOST_PORT:-32886}/
  Use HTTP, not HTTPS. Allow TCP port ${OBSHELL_HOST_PORT:-32886} in the host firewall/security group.

Monitoring is internal to the Compose network by default.
For optional host-loopback access, start with compose.monitoring.yaml and use
the same overlay for every subsequent Compose command.
  Prometheus: http://127.0.0.1:${PROMETHEUS_HOST_PORT:-39090}/
  Alertmanager: http://127.0.0.1:${ALERTMANAGER_HOST_PORT:-39093}/

Accounts: root for obshell, root@sys for SQL.
Retrieve the root password on the Docker host, from the deployment directory:
  sudo docker compose -p ${compose_project} run --rm --no-deps observer-bootstrap --mode show-password
If startup used -f or --env-file, include those same options before run.
If Docker is accessible without sudo, you may omit sudo.
This prints the root password; do not share its output.
EOF
fi
