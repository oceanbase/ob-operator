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

  cat <<EOF

Lakehouse deployment completed.

OceanBase SQL:
  ${host_ip}:${OBSERVER_SQL_HOST_PORT:-32881}

obshell Dashboard:
  http://${host_ip}:${OBSHELL_HOST_PORT:-32886}/
  Use HTTP, not HTTPS. Allow TCP port ${OBSHELL_HOST_PORT:-32886} in the host firewall/security group.

Prometheus (available when started):
  http://${host_ip}:${PROMETHEUS_HOST_PORT:-39090}/

Alertmanager (available when started):
  http://${host_ip}:${ALERTMANAGER_HOST_PORT:-39093}/

Login user: root
Retrieve the generated root@sys password:
  docker compose run --rm --no-deps observer-bootstrap --mode show-password
EOF
fi
