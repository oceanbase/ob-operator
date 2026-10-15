#!/usr/bin/env bash
# One command, three resumable stages; never rolls back or deletes persistent data.
set -euo pipefail
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
helm_args=(--history-max 10)
while (( $# )); do
  case "$1" in
    -f|--values|--set|--set-string|--set-file)
      (( $# >= 2 )) || { echo 'Missing Helm option value.' >&2; exit 2; }
      helm_args+=("$1" "$2"); shift 2 ;;
    *) echo "Unsupported option: $1. Only values flags are accepted; atomic/force/cleanup are intentionally disallowed." >&2; exit 2 ;;
  esac
done
timeout=${LAKEHOUSE_TIMEOUT:-1800}
[[ $timeout =~ ^[1-9][0-9]{0,4}$ ]] || { echo 'LAKEHOUSE_TIMEOUT must be 1-99999 seconds.' >&2; exit 2; }
# Fail before creating resources if this is not an SS-enabled Operator cluster.
for crd in obclusters oblogserviceclusters oblogservicezones oblogservicenodes; do
  kubectl get "crd/$crd.oceanbase.oceanbase.com" >/dev/null
  kubectl wait "crd/$crd.oceanbase.oceanbase.com" --for=condition=Established --timeout=30s >/dev/null
done
if ! kubectl get namespace "$namespace" >/dev/null 2>&1; then
  kubectl create namespace "$namespace" >/dev/null
fi
diagnostics() {
  echo "Installation stopped; data has NOT been automatically removed. Resume with the same command after resolving the failure." >&2
  kubectl -n "$namespace" get pods,jobs,obclusters,oblogserviceclusters >&2 || true
  # Do not print rendered manifests, Secret data, or raw Operator logs.
}
trap diagnostics ERR
releases=$(helm list -n "$namespace" --all -o json)
stage=storage
if jq -e --arg name "$release" '.[] | select(.name == $name)' <<<"$releases" >/dev/null; then
  # No --reuse-values: the caller supplies the same complete values file on every
  # invocation. Lookup-based identity protection rejects accidental drift.
  stage=$(helm get values "$release" -n "$namespace" --all -o json | jq -er '.stage')
fi
# Helm uninstall retains CRs. Reinstall must not temporarily drop to storage stage.
if kubectl -n "$namespace" get "obcluster/$release-ob" >/dev/null 2>&1; then
  stage=cluster
elif kubectl -n "$namespace" get "oblogservicecluster/$release-ls" >/dev/null 2>&1; then
  stage=logservice
fi
case "$stage" in storage|logservice|cluster) ;; *) echo 'Unknown retained installation stage.' >&2; exit 1 ;; esac
upgrade() {
  helm upgrade --install "$release" "$chart_dir" -n "$namespace" \
    "${helm_args[@]}" --set-string "stage=$1" --wait --wait-for-jobs --timeout "${timeout}s"
}
wait_running() {
  kubectl -n "$namespace" wait "$1" --for=jsonpath='{.status.status}'=running --timeout="${timeout}s"
}
if [[ $stage == storage ]]; then
  upgrade storage
  stage=logservice
fi
if [[ $stage == logservice ]]; then
  upgrade logservice
  wait_running "oblogservicecluster/$release-ls"
fi
upgrade cluster
wait_running "oblogservicecluster/$release-ls"
wait_running "obcluster/$release-ob"
kubectl -n "$namespace" wait pods -l "ref-obcluster=$release-ob" --for=condition=Ready --timeout="${timeout}s"
echo "Lakehouse ready: namespace=$namespace LogService=$release-ls OBCluster=$release-ob"
echo 'Database root password is in the configured Secret (default: RELEASE-db-root, key: password).'
echo 'Use kubectl get observers and the per-observer Services for SQL connectivity; no public ports are opened automatically.'
