#!/usr/bin/env bash
# One command, resumable stages; never rolls back or deletes persistent data.
set -Eeuo pipefail
export LC_ALL=C
if (( $# < 2 )); then
  echo "Usage: bash install.sh RELEASE NAMESPACE [-f values.yaml] [--set key=value]" >&2
  exit 2
fi
release=$1
namespace=$2
shift 2
[[ $release =~ ^[a-z][a-z0-9-]{0,38}[a-z0-9]$ || $release =~ ^[a-z]$ ]] || { echo 'Invalid release name (1-40 DNS characters).' >&2; exit 2; }
[[ $namespace =~ ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$ ]] || { echo 'Invalid namespace.' >&2; exit 2; }
for tool in helm kubectl jq; do command -v "$tool" >/dev/null || { echo "Missing tool: $tool" >&2; exit 2; }; done
chart_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
user_args=()
while (( $# )); do
  case "$1" in
    -f|--values|--set|--set-string|--set-file)
      (( $# >= 2 )) || { echo 'Missing Helm option value.' >&2; exit 2; }
      user_args+=("$1" "$2"); shift 2 ;;
    *) echo "Unsupported option: $1. Only values flags are accepted; atomic/force/cleanup are intentionally disallowed." >&2; exit 2 ;;
  esac
done
timeout=${LAKEHOUSE_TIMEOUT:-1800}
request_timeout=${LAKEHOUSE_REQUEST_TIMEOUT:-30}
diagnostic_timeout=${LAKEHOUSE_DIAGNOSTIC_TIMEOUT:-10}
progress_interval=${LAKEHOUSE_PROGRESS_INTERVAL:-30}
for setting in timeout request_timeout diagnostic_timeout progress_interval; do
  [[ ${!setting} =~ ^[1-9][0-9]{0,4}$ ]] || { echo "$setting must be 1-99999 seconds." >&2; exit 2; }
done
umask 077
work_dir=$(mktemp -d "${TMPDIR:-/tmp}/oceanbase-lakehouse-${release}.XXXXXX")
log_file=$work_dir/install.log
: > "$log_file"
base_values=$work_dir/previous-values.json
started=$SECONDS
current_action=preflight
active_pid=''
monitor_pid=''
echo "Detailed log: $log_file (private, mode 0600)"

# Helm's read commands have no --timeout flag. Bound the executable itself as
# well as kubectl requests. The monitor also provides progress for quiet Helm
# waits. It never prints command arguments (which can contain credentials).
run_bounded() {
  local budget=$1 label=$2 command_pid watcher_pid marker result
  shift 2
  marker=$(mktemp "$work_dir/deadline.XXXXXX")
  rm -f "$marker"
  "$@" 2>>"$log_file" &
  command_pid=$!
  active_pid=$command_pid
  (
    ticker_pid=''
    trap '[[ -z $ticker_pid ]] || kill "$ticker_pid" 2>/dev/null || true; exit 0' TERM INT
    local_begin=$SECONDS
    next_progress=$progress_interval
    while kill -0 "$command_pid" 2>/dev/null; do
      elapsed=$((SECONDS - local_begin))
      if (( elapsed >= budget )); then
        : > "$marker"
        echo "Timed out after ${budget}s: $label" >&2
        kill -TERM "$command_pid" 2>/dev/null || true
        /bin/sleep 1 & ticker_pid=$!; wait "$ticker_pid" || true
        kill -KILL "$command_pid" 2>/dev/null || true
        exit 0
      fi
      if (( elapsed >= next_progress )); then
        echo "[wait] $label: ${elapsed}s elapsed (limit ${budget}s)" >&2
        next_progress=$((elapsed + progress_interval))
      fi
      /bin/sleep 1 & ticker_pid=$!; wait "$ticker_pid" || true
    done
  ) &
  watcher_pid=$!
  monitor_pid=$watcher_pid
  if wait "$command_pid"; then result=0; else result=$?; fi
  kill "$watcher_pid" 2>/dev/null || true
  wait "$watcher_pid" 2>/dev/null || true
  active_pid=''
  monitor_pid=''
  if [[ -f $marker ]]; then result=124; fi
  rm -f "$marker"
  return "$result"
}

# Summarize only structured status/reason fields. Raw Events messages, Pod specs,
# rendered manifests, Secret data and Operator logs never reach the terminal.
diagnostics() {
  local resources events
  echo "Installation stopped during $current_action; data has NOT been automatically removed." >&2
  echo "Resume with the same release and namespace after resolving the failure; saved Helm user values will be reused, and new values flags override them." >&2
  echo "Private installation log: $log_file" >&2
  if resources=$(run_bounded "$diagnostic_timeout" 'resource diagnostics' kubectl -n "$namespace" get pods,jobs,obclusters,oblogserviceclusters,pvc -o json --request-timeout="${diagnostic_timeout}s"); then
    jq -r '.items[] | [.kind, .metadata.name, (.status.phase // .status.status // ""), ([.status.containerStatuses[]?.state.waiting.reason] | join(","))] | @tsv' <<<"$resources" >&2 || true
  fi
  if events=$(run_bounded "$diagnostic_timeout" 'warning event diagnostics' kubectl -n "$namespace" get events --field-selector=type=Warning -o json --request-timeout="${diagnostic_timeout}s"); then
    jq -r '.items | sort_by(.lastTimestamp // .metadata.creationTimestamp) | .[-10:][] | ["Warning", .reason, .involvedObject.kind, .involvedObject.name] | @tsv' <<<"$events" >&2 || true
    if jq -e '.items[] | select(.reason == "FailedScheduling")' <<<"$events" >/dev/null; then
      echo 'Check node capacity, resource requests, taints and volume binding for FailedScheduling.' >&2
    fi
    if jq -e '.items[] | select(.reason == "FailedMount" or .reason == "FailedBinding" or .reason == "ProvisioningFailed")' <<<"$events" >/dev/null; then
      echo 'Check StorageClass, PVC status and storage provisioner availability.' >&2
    fi
  fi
}
finish() {
  local result=$?
  trap - EXIT INT TERM
  set +e
  [[ -z $active_pid ]] || kill "$active_pid" 2>/dev/null
  [[ -z $monitor_pid ]] || kill "$monitor_pid" 2>/dev/null
  if (( result != 0 )); then diagnostics; fi
  rm -f "$base_values" "$base_values.raw"
  exit "$result"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

phase_label() {
  # The package entry point already numbers its five overall steps. Keep the
  # six detailed phases named, rather than starting a competing progress count.
  if [[ ${LAKEHOUSE_MANAGED_ACCESS:-0} == 1 ]]; then
    printf '%s' "${1#*] }"
  else
    printf '%s' "$1"
  fi
}
announce() {
  current_action=$(phase_label "$1")
  echo "[stage] $current_action ($((SECONDS - started))s elapsed)"
}
completed() {
  echo "[done] $(phase_label "$1") ($((SECONDS - started))s elapsed)"
}
skipped() {
  echo "[skip] $(phase_label "$1")"
}
announce '[1/6] Preflight: checking SS-enabled CRDs and existing release'
for crd in obclusters oblogserviceclusters oblogservicezones oblogservicenodes; do
  run_bounded "$request_timeout" "read CRD $crd" kubectl get "crd/$crd.oceanbase.oceanbase.com" --request-timeout="${request_timeout}s" >>"$log_file"
  run_bounded "$request_timeout" "wait for CRD $crd" kubectl wait "crd/$crd.oceanbase.oceanbase.com" --for=condition=Established --timeout="${request_timeout}s" --request-timeout="${request_timeout}s" >>"$log_file"
done
# NotFound is different from Forbidden/network failure. Never create or advance
# an installation after an unsuccessful read of existing state.
namespace_resource=$(run_bounded "$request_timeout" 'read namespace' kubectl get namespace "$namespace" --ignore-not-found -o name --request-timeout="${request_timeout}s")
if [[ -z $namespace_resource ]]; then
  run_bounded "$request_timeout" 'create namespace' kubectl create namespace "$namespace" --request-timeout="${request_timeout}s" >>"$log_file"
fi
releases=$(run_bounded "$request_timeout" 'read Helm releases' helm list -n "$namespace" --all --filter "^${release}$" -o json)
stage=storage
helm_args=(--history-max 10)
if jq -e --arg name "$release" '.[] | select(.name == $name)' <<<"$releases" >/dev/null; then
  # Only explicit values from the installed release are the base: new chart
  # defaults still apply. JSON is valid YAML. Never print these secret-bearing
  # values or retain them after exit; caller flags always take precedence.
  run_bounded "$request_timeout" 'read installed user values' helm get values "$release" -n "$namespace" -o json > "$base_values.raw"
  jq 'if . == null then {} elif type == "object" then . else error("invalid installed user values") end' "$base_values.raw" > "$base_values"
  stage=$(jq -er '.stage // "storage"' "$base_values")
  helm_args+=(-f "$base_values")
  echo 'Reusing installed user values; explicit values flags take precedence.'
fi
helm_args+=("${user_args[@]}")
# Helm uninstall retains CRs. Reinstall must not temporarily drop to storage stage.
retained_cluster=$(run_bounded "$request_timeout" 'read retained OBCluster' kubectl -n "$namespace" get "obcluster/$release-ob" --ignore-not-found -o name --request-timeout="${request_timeout}s")
retained_logservice=$(run_bounded "$request_timeout" 'read retained LogService' kubectl -n "$namespace" get "oblogservicecluster/$release-ls" --ignore-not-found -o name --request-timeout="${request_timeout}s")
if [[ -n $retained_cluster ]]; then stage=cluster
elif [[ -n $retained_logservice ]]; then stage=logservice
fi
case "$stage" in storage|logservice|cluster) ;; *) echo 'Unknown retained installation stage.' >&2; exit 1 ;; esac
completed '[1/6] Preflight: existing installation state checked'
upgrade() {
  run_bounded "$timeout" "Helm $1 stage" helm upgrade --install "$release" "$chart_dir" -n "$namespace" \
    "${helm_args[@]}" --set-string "stage=$1" --wait --wait-for-jobs --timeout "${timeout}s" >>"$log_file"
}
wait_running() {
  local deadline=$((SECONDS + timeout)) began=$SECONDS remaining status previous='' delay budget next_progress=$SECONDS quiet=${2:-false}
  current_action="Wait for $1 to become running"
  while (( SECONDS < deadline )); do
    remaining=$((deadline - SECONDS))
    (( remaining > 0 )) || break
    budget=$request_timeout
    (( remaining >= budget )) || budget=$remaining
    status=$(run_bounded "$budget" "read $1 status" kubectl -n "$namespace" get "$1" -o jsonpath='{.status.status}' --request-timeout="${budget}s")
    case "$status" in
      running)
        if [[ $quiet != true ]]; then echo "[ready] $1 is running ($((SECONDS - began))s)"; fi
        return 0 ;;
      failed)
        echo "$1 entered failed status. Inspect the Operator events and resolve the cause before retrying; no data is deleted or bootstrap forced." >&2
        return 1 ;;
    esac
    # A previously ready LogService is still rechecked after the cluster apply.
    # Stay quiet only if it remains ready; otherwise show recovery progress.
    quiet=false
    # Normalize before comparing, so an unknown API value does not cause an
    # identical 'pending' message on every poll (and never exposes raw API text).
    case "$status" in new|pending|creating|initializing|upgrading|running|failed) ;; *) status=pending ;; esac
    if [[ $status != "$previous" ]] || (( SECONDS >= next_progress )); then
      # Do not echo arbitrary API text. Known status values aid progress; unknown
      # values stay in the API and are summarized without exposing their content.
      echo "[wait] $1: status=$status; $((SECONDS - began))s elapsed (limit ${timeout}s)"
      previous=$status
      next_progress=$((SECONDS + progress_interval))
    fi
    remaining=$((deadline - SECONDS))
    (( remaining > 0 )) || break
    delay=5
    (( remaining >= delay )) || delay=$remaining
    sleep "$delay"
  done
  echo "Timed out waiting for $1 to become running." >&2
  return 1
}
logservice_reported=false
if [[ $stage == storage ]]; then
  announce '[2/6] Storage: apply resources and wait for initialization'
  upgrade storage
  completed '[2/6] Storage: initialization Job and Helm resources ready'
  stage=logservice
else
  skipped '[2/6] Storage: resume past this stage; Helm will still check retained resources'
fi
if [[ $stage == logservice ]]; then
  announce '[3/6] LogService: apply configuration and wait for running status'
  upgrade logservice
  wait_running "oblogservicecluster/$release-ls"
  logservice_reported=true
  completed '[3/6] LogService: running status confirmed'
else
  skipped '[3/6] LogService: resume at cluster stage; final status is verified below'
fi
announce '[4/6] OceanBase: apply cluster configuration and verify database status'
upgrade cluster
wait_running "oblogservicecluster/$release-ls" "$logservice_reported"
wait_running "obcluster/$release-ob"
completed '[4/6] OceanBase: LogService and OBCluster running status confirmed'
announce '[5/6] Observer Pods: wait for Ready conditions'
run_bounded "$timeout" 'Observer Pod readiness' kubectl -n "$namespace" wait pods -l "ref-obcluster=$release-ob" --for=condition=Ready --timeout="${timeout}s" --request-timeout="${timeout}s" >>"$log_file"
completed '[5/6] Observer Pods: Ready conditions confirmed'
# Read resolved values without printing credentials from Helm's stored values.
announce '[6/6] Dashboard: check configuration and HTTP readiness probes'
dashboard_config=$(run_bounded "$request_timeout" 'read Dashboard configuration' helm get values "$release" -n "$namespace" --all -o json |
  jq -c '.dashboard | {enabled, port: .service.port}')
dashboard_enabled=$(jq -r '.enabled' <<<"$dashboard_config")
case "$dashboard_enabled" in
  true)
    run_bounded "$timeout" 'Dashboard rollout' kubectl -n "$namespace" rollout status "deployment/oceanbase-dashboard-$release" --timeout="${timeout}s" --request-timeout="${timeout}s" >>"$log_file"
    completed '[6/6] Dashboard: Deployment and its HTTP readiness probes ready'
    ;;
  false) skipped '[6/6] Dashboard: disabled by configuration' ;;
  *) echo 'Cannot determine Dashboard readiness: resolved dashboard.enabled is missing.' >&2; exit 1 ;;
esac
echo "Lakehouse ready: namespace=$namespace LogService=$release-ls OBCluster=$release-ob ($((SECONDS - started))s)"
# The package entry point verifies HTTP access and prints one combined summary.
# Only suppress expert access instructions after every readiness gate has passed.
if [[ ${LAKEHOUSE_MANAGED_ACCESS:-0} == 1 ]]; then
  exit 0
fi
echo 'Database root password is in the configured Secret (default: RELEASE-db-root, key: password).'
echo 'Use kubectl get observers and the per-observer Services for SQL connectivity; no public database ports are opened automatically.'
if [[ $dashboard_enabled == true ]]; then
  dashboard_port=$(jq -er '.port' <<<"$dashboard_config")
  echo "Dashboard ready: oceanbase-dashboard-$release"
  echo "  kubectl -n $namespace port-forward svc/oceanbase-dashboard-$release 8080:$dashboard_port"
  echo 'Then open http://127.0.0.1:8080 on the computer running port-forward.'
  echo 'Login: admin. Retrieve the INITIAL password (use your new password after changing it):'
  echo "  kubectl -n $namespace get secret $release-dashboard-bootstrap -o jsonpath='{.data.password}' | base64 --decode; echo"
else
  echo 'Dashboard is disabled (dashboard.enabled=false).'
fi
