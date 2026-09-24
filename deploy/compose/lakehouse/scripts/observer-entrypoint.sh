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
if [[ -f "${backup_config}" ]]; then
  mkdir -p "${home}/etc"
  cp "${backup_config}" "${runtime_config}"
  exec "${home}/bin/observer" --nodaemon
fi

options="config_additional_dir=${home}/store/etc,cpu_count=${CPU_COUNT},datafile_size=${DATAFILE_SIZE},enable_syslog_recycle=true,max_syslog_file_count=4,enable_logservice=True,memory_limit=${MEMORY_LIMIT}"
exec "${home}/bin/observer" \
  -m shared_storage \
  --nodaemon \
  --appname "${CLUSTER_NAME}" \
  --cluster_id "${CLUSTER_ID}" \
  --zone "${ZONE_NAME}" \
  -I "${LOCAL_IP}" \
  -p 2881 \
  -P 2882 \
  -d "${data_dir}" \
  -l info \
  -o "${options}"
