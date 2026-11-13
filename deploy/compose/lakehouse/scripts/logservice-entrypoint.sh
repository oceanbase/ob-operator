#!/usr/bin/env bash
set -euo pipefail

home=/home/admin/oblogservice
store_dir="${home}/store"
log_dir="${home}/log"
persist_etc="${store_dir}/etc"
local_storage="${store_dir}/data"

mkdir -p "${persist_etc}" "${local_storage}" "${log_dir}"
if [[ ! -f "${persist_etc}/config.toml" ]]; then
  cp "${home}"/etc/*.yml "${persist_etc}/" 2>/dev/null || true
fi
rm -rf "${home:?}/etc"
ln -s "${persist_etc}" "${home}/etc"

args=()
if [[ ! -f "${persist_etc}/config.toml" ]]; then
  args=(-g "cluster_id=${CLUSTER_ID},local_ip=${LOCAL_IP},rpc_port=50051,http_port=50052,local_storage_dir=${local_storage},log_disk_size=${LOG_DISK_SIZE},memory_limit=${MEMORY_LIMIT}")
fi

"${home}/bin/oblogservice" "${args[@]}"

# Procfs is part of the Linux runtime; do not depend on optional procps/pgrep.
find_daemon_pid() {
  local stat_file stat_line process_state candidate oldest=""
  [[ -r /proc/self/stat ]] || { echo "cannot inspect /proc/self/stat" >&2; return 1; }
  for stat_file in /proc/[0-9]*/stat; do
    IFS= read -r stat_line < "${stat_file}" 2>/dev/null || continue
    [[ "${stat_line#* (}" == oblogservice\)\ * ]] || continue
    process_state="${stat_line##*) }"
    process_state="${process_state%% *}"
    [[ "${process_state}" != Z && "${process_state}" != X ]] || continue
    candidate="${stat_file%/stat}"
    candidate="${candidate##*/}"
    if [[ -z "${oldest}" ]] || (( candidate < oldest )); then oldest="${candidate}"; fi
  done
  printf '%s' "${oldest}"
}

deadline=$((SECONDS + 300))
pid=""
while (( SECONDS < deadline )); do
  pid="$(find_daemon_pid)" || { echo "failed to inspect LogService processes" >&2; exit 1; }
  [[ -n "${pid}" ]] && break
  sleep 1
done
if [[ -z "${pid}" ]]; then
  echo "oblogservice daemon did not start" >&2
  exit 1
fi

daemon_running() {
  local stat_line state
  IFS= read -r stat_line < "/proc/${pid}/stat" 2>/dev/null || return 1
  state="${stat_line##*) }"; state="${state%% *}"
  [[ "$state" != Z && "$state" != X ]]
}
stop_daemon() {
  trap '' TERM INT
  echo "Stopping LogService pid ${pid}" >&2
  kill -TERM "${pid}" 2>/dev/null || true
  local stop_deadline=$((SECONDS + 20))
  while daemon_running && (( SECONDS < stop_deadline )); do sleep 0.1; done
  if daemon_running; then
    echo "WARNING: LogService did not stop within 20s; forcing SIGKILL" >&2
    kill -KILL "${pid}" 2>/dev/null || true
    stop_deadline=$((SECONDS + 3))
    while daemon_running && (( SECONDS < stop_deadline )); do sleep 0.1; done
  fi
  if daemon_running; then echo "LogService remains alive after SIGKILL" >&2; exit 1; fi
  exit 0
}
trap stop_daemon TERM INT

while daemon_running; do
  sleep 5 &
  wait $! || true
done
wait "${pid}" 2>/dev/null || true
echo "oblogservice daemon exited" >&2
exit 1
