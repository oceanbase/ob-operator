#!/usr/bin/env bash
set -euo pipefail

home=/home/admin/oceanbase
data_dir=/home/admin/data-file
log_dir=/home/admin/log

mkdir -p "${log_dir}/log" "${home}/store" \
  "${data_dir}/slog" "${data_dir}/etc" "${data_dir}/sort_dir" \
  "${data_dir}/sstable" "${data_dir}/wallet"
ln -sfn "${log_dir}/log" "${home}/log"
ln -sfn "${data_dir}/slog" "${home}/store/slog"
ln -sfn "${data_dir}/etc" "${home}/store/etc"
ln -sfn "${data_dir}/sort_dir" "${home}/store/sort_dir"
ln -sfn "${data_dir}/sstable" "${home}/store/sstable"
ln -sfn "${data_dir}/wallet" "${home}/wallet"

backup_config="${home}/store/etc/observer.conf.bin"
runtime_config="${home}/etc/observer.config.bin"
observer_args=(--nodaemon)
if [[ -f "${backup_config}" ]]; then
  mkdir -p "${home}/etc"
  cp "${backup_config}" "${runtime_config}"
else
  options="config_additional_dir=${home}/store/etc,cpu_count=${CPU_COUNT},datafile_size=${DATAFILE_SIZE},datafile_maxsize=${DATAFILE_MAXSIZE:-47G},datafile_next=${DATAFILE_NEXT:-5G},enable_syslog_recycle=true,max_syslog_file_count=4,enable_logservice=True,memory_limit=${MEMORY_LIMIT},__min_full_resource_pool_memory=2147483648"
  observer_args+=(
    -m shared_storage
    --appname "${CLUSTER_NAME}"
    --cluster_id "${CLUSTER_ID}"
    --zone "${ZONE_NAME}"
    -I "${LOCAL_IP}"
    -p 2881
    -P 2882
    -d "${data_dir}"
    -l info
    -o "${options}"
  )
fi

"${home}/bin/observer" "${observer_args[@]}" &
observer_pid=$!
# --nodaemon does not refresh a PID file inherited from the base image.
# obshell uses this file to determine availability even when SQL is healthy.
mkdir -p "${home}/run"
printf '%s\n' "${observer_pid}" > "${home}/run/observer.pid"
obshell_refresh_pid=""
monitoring_register_pid=""

stop_processes() {
  trap - TERM INT
  if [[ -n "${obshell_refresh_pid}" ]] && kill -0 "${obshell_refresh_pid}" 2>/dev/null; then
    kill -TERM "${obshell_refresh_pid}" 2>/dev/null || true
    wait "${obshell_refresh_pid}" 2>/dev/null || true
  fi
  if [[ -n "${monitoring_register_pid}" ]] && kill -0 "${monitoring_register_pid}" 2>/dev/null; then
    kill -TERM "${monitoring_register_pid}" 2>/dev/null || true
    wait "${monitoring_register_pid}" 2>/dev/null || true
  fi
  "${home}/bin/obshell" agent stop >/dev/null 2>&1 || true
  if kill -0 "${observer_pid}" 2>/dev/null; then
    kill -TERM "${observer_pid}" 2>/dev/null || true
    wait "${observer_pid}" 2>/dev/null || true
  fi
}
trap 'stop_processes; exit 143' TERM INT

for _ in $(seq 1 180); do
  if ! kill -0 "${observer_pid}" 2>/dev/null; then
    wait "${observer_pid}"
    exit $?
  fi
  if bash -c 'exec 3<>/dev/tcp/127.0.0.1/2881' 2>/dev/null; then
    break
  fi
  sleep 1
done
if ! bash -c 'exec 3<>/dev/tcp/127.0.0.1/2881' 2>/dev/null; then
  echo "Observer did not open SQL port before the obshell startup deadline" >&2
  stop_processes
  exit 1
fi

root_password="$(cat /run/secrets/ob_root_password)"
[[ -n "${root_password}" ]] || {
  echo "Observer root password credential is empty" >&2
  stop_processes
  exit 1
}
obshell_args=(agent start --ip "${LOCAL_IP}" --password "${root_password}")
start_obshell() {
  "${home}/bin/obshell" agent stop >/dev/null 2>&1 || true
  rm -f "${home}/run/daemon.pid"
  "${home}/bin/obshell" "${obshell_args[@]}"
}

bootstrap_marker=/run/secrets/observer_bootstrap_complete
bootstrap_pending=false
[[ -f "${bootstrap_marker}" ]] || bootstrap_pending=true
if ! start_obshell; then
  echo "obshell agent failed to start" >&2
  stop_processes
  exit 1
fi

if [[ "${AUTO_REGISTER_MONITORING:-true}" == true ]]; then
  /opt/lakehouse/obshell-monitoring-register.sh &
  monitoring_register_pid=$!
fi

if [[ "${bootstrap_pending}" == true ]]; then
  (
    while [[ ! -f "${bootstrap_marker}" ]]; do
      kill -0 "${observer_pid}" 2>/dev/null || exit 1
      sleep 2
    done
    if ! start_obshell; then
      echo "obshell could not take over the initialized cluster" >&2
      kill -TERM "${observer_pid}" 2>/dev/null || true
      exit 1
    fi
  ) &
  obshell_refresh_pid=$!
fi

status=0
wait "${observer_pid}" || status=$?
"${home}/bin/obshell" agent stop >/dev/null 2>&1 || true
exit "${status}"
