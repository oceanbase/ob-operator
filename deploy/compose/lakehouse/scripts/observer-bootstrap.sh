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
  credential_source="$(shell_quote "${HOSTNAME}:/run/secrets/ob_root_password")"
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
Retrieve the saved root password on the Docker host (any directory):
  sudo docker cp ${credential_source} - | tar -xOf -
This command reads this bootstrap container's saved credential, even after it exits.
No Compose files, project flags or environment options are needed. Keep this container until retrieval.
If Docker is accessible without sudo, you may omit sudo.
This reads the saved credential file, not the current password from the database.
After changing the password manually, use the new password; the saved file may be stale.
This prints a password; do not share its output.
EOF
fi
