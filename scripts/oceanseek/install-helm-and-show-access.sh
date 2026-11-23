#!/usr/bin/env bash
# Install a oceanseek Chart, then print addresses after Kubernetes allocates ports.
set -euo pipefail

if (( $# < 2 )); then
  echo 'Usage: install-helm-and-show-access.sh RELEASE CHART [helm upgrade --install options]' >&2
  exit 2
fi

release=$1
chart=$2
shift 2
namespace=default
args=("$@")
kubectl_args=()
for ((i = 0; i < ${#args[@]}; i++)); do
  case "${args[i]}" in
    -n|--namespace)
      if (( i + 1 >= ${#args[@]} )); then
        echo 'Missing namespace after -n/--namespace' >&2
        exit 2
      fi
      namespace=${args[i+1]}
      ((i += 1))
      ;;
    --namespace=*) namespace=${args[i]#*=} ;;
    --kubeconfig|--kube-context)
      if (( i + 1 >= ${#args[@]} )); then echo "Missing value for ${args[i]}" >&2; exit 2; fi
      option=${args[i]}; [[ $option != --kube-context ]] || option=--context
      kubectl_args+=("$option" "${args[i+1]}"); ((i += 1)) ;;
    --kubeconfig=*) kubectl_args+=("${args[i]}") ;;
    --kube-context=*) kubectl_args+=("--context=${args[i]#*=}") ;;
  esac
done

if helm upgrade --install "$release" "$chart" "$@"; then
  :
else
  status=$?
  echo 'Helm installation failed. Check storage initialization first; its logs identify the bucket and Secret reference if storage access failed:' >&2
  printf '  kubectl' >&2
  printf ' %q' "${kubectl_args[@]}" -n "$namespace" logs -l "app.kubernetes.io/instance=$release,app.kubernetes.io/component=storage-init" --all-containers=true --prefix=true --tail=50 >&2
  printf '\n' >&2
  echo 'If no initialization logs exist, inspect Pod events for missing Secrets, image pulls, or scheduling failures.' >&2
  exit "$status"
fi

kubectl() { command kubectl "${kubectl_args[@]}" "$@"; }

helm_context=()
for ((i=0; i<${#kubectl_args[@]}; i++)); do
  option=${kubectl_args[i]}; option=${option/--context/--kube-context}; helm_context+=("$option")
done
configured_host=$(helm get values "$release" -n "$namespace" "${helm_context[@]}" -o json | python3 -c 'import json,sys; print(((json.load(sys.stdin) or {}).get("access") or {}).get("host", ""))')
configured_host=${OCEANSEEK_ACCESS_HOST:-$configured_host}
if [[ -n $configured_host && ! $configured_host =~ ^[A-Za-z0-9._:-]+$ ]]; then
  echo 'Helm succeeded, but the node address contains unsupported characters.' >&2
  exit 1
fi

node_address_for() {
  if [[ -n $configured_host ]]; then
    printf '%s' "$configured_host"
    return
  fi
  local families
  families=$(kubectl -n "$namespace" get service "$1" -o 'jsonpath={.spec.ipFamilies[*]}') || return
  kubectl get nodes -o json | python3 -c '
import json, sys
families = sys.argv[1].split()
nodes = json.load(sys.stdin)["items"]
addresses = [a for node in nodes
             if any(c["type"] == "Ready" and c["status"] == "True" for c in node["status"].get("conditions", []))
             for a in node["status"].get("addresses", [])]
for kind in ("ExternalIP", "InternalIP"):
    for address in addresses:
        family = "IPv6" if ":" in address["address"] else "IPv4"
        if address["type"] == kind and (not families or family in families):
            print(address["address"])
            sys.exit(0)
sys.exit("No suitable Ready node address was found. Set access.host to a reachable node IP or DNS name.")
' "$families"
}

dashboard_service="oceanbase-dashboard-$release"
if kubectl -n "$namespace" get service "$dashboard_service" >/dev/null 2>&1; then
  dashboard_type=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.spec.type}')
  service_port=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.spec.ports[0].port}')
  if [[ -z $service_port ]]; then echo 'Dashboard Service port is unavailable.' >&2; exit 1; fi
  case $dashboard_type in
    NodePort)
      url_host=$(node_address_for "$dashboard_service")
      if [[ $url_host == *:* ]]; then url_host="[$url_host]"; fi
      dashboard_port=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.spec.ports[0].nodePort}')
      if [[ -z $dashboard_port ]]; then echo 'Dashboard NodePort is not allocated.' >&2; exit 1; fi
      printf 'Dashboard URL: http://%s:%s/#/login\n' "$url_host" "$dashboard_port"
      ;;
    LoadBalancer)
      dashboard_address=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
      if [[ -z $dashboard_address ]]; then
        dashboard_address=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
      fi
      if [[ -n $dashboard_address ]]; then
        if [[ $dashboard_address == *:* ]]; then dashboard_address="[$dashboard_address]"; fi
        printf 'Dashboard URL: http://%s:%s/#/login\n' "$dashboard_address" "$service_port"
      else
        echo 'Dashboard LoadBalancer address is still pending.'
      fi
      ;;
    *)
      printf '\nOn the computer running your browser, start:\n  kubectl'
      printf ' %q' "${kubectl_args[@]}" -n "$namespace" port-forward "svc/$dashboard_service" "8080:$service_port"
      printf '\nOpen http://127.0.0.1:8080/#/login\n'
      ;;
  esac
else
  echo 'Dashboard Service is absent; check dashboard.enabled in the release values.'
fi

sql_service="$release-sql"
if kubectl -n "$namespace" get service "$sql_service" >/dev/null 2>&1; then
  sql_type=$(kubectl -n "$namespace" get service "$sql_service" -o jsonpath='{.spec.type}')
  if [[ $sql_type == NodePort ]]; then
    node_host=$(node_address_for "$sql_service")
    sql_port=$(kubectl -n "$namespace" get service "$sql_service" -o jsonpath='{.spec.ports[0].nodePort}')
    if [[ -z $sql_port ]]; then echo 'SQL NodePort is not allocated.' >&2; exit 1; fi
    printf "SQL command: mysql -h '%s' -P %s -u 'root@sys' --protocol=TCP -p\n" "$node_host" "$sql_port"
  else
    sql_port=$(kubectl -n "$namespace" get service "$sql_service" -o jsonpath='{.spec.ports[0].port}')
    if [[ -z $sql_port ]]; then echo 'SQL Service port is unavailable.' >&2; exit 1; fi
    printf 'For local SQL access, run: kubectl -n %s port-forward svc/%s 2883:%s\n' "$namespace" "$sql_service" "$sql_port"
    printf "SQL command: mysql -h 127.0.0.1 -P 2883 -u 'root@sys' -p\n"
  fi
else
  echo 'SQL Service is absent; check stage=cluster and the release status.'
fi
echo 'Use the password retrieval commands from Helm NOTES in a trusted terminal.'
