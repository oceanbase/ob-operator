#!/bin/sh
set -eu

host_ip="${HOST_ACCESS_IP:-}"
if [ -z "${host_ip}" ]; then
  # This container uses host networking, so the route source is the Docker
  # host's address rather than a Compose bridge address.
  set -- $(ip route get "${HOST_IP_PROBE:-1.1.1.1}" 2>/dev/null)
  while [ "$#" -gt 1 ]; do
    if [ "$1" = src ]; then
      host_ip="$2"
      break
    fi
    shift
  done
fi

case "${host_ip}" in
  ''|*[!0-9A-Fa-f:.]*)
    echo "Unable to determine Docker host IP; set HOST_ACCESS_IP in .env" >&2
    exit 1
    ;;
esac

target=/run/secrets/docker_host_ip
temporary="$(mktemp "${target}.tmp.XXXXXX")"
trap 'rm -f "${temporary}"' EXIT HUP INT TERM
printf '%s\n' "${host_ip}" >"${temporary}"
chmod 0600 "${temporary}"
mv -f "${temporary}" "${target}"
trap - EXIT HUP INT TERM
echo "Docker host access IP: ${host_ip}"
