#!/usr/bin/env bash
# Install a lakehouse Chart, then print addresses after Kubernetes allocates ports.
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

helm upgrade --install "$release" "$chart" "$@"

kubectl() { command kubectl "${kubectl_args[@]}" "$@"; }

helm_context=()
for ((i=0; i<${#kubectl_args[@]}; i++)); do
  option=${kubectl_args[i]}; option=${option/--context/--kube-context}; helm_context+=("$option")
done
configured_host=$(helm get values "$release" -n "$namespace" "${helm_context[@]}" -o json | python3 -c 'import json,sys; print(((json.load(sys.stdin) or {}).get("access") or {}).get("host", ""))')
if [[ -n ${LAKEHOUSE_ACCESS_HOST:-} ]]; then
  node_host=$LAKEHOUSE_ACCESS_HOST
elif [[ -n $configured_host ]]; then
  node_host=$configured_host
else
  nodes=$(kubectl get nodes -o wide --no-headers)
  node_host=$(awk '$2 ~ /^Ready/ {print $6; exit}' <<< "$nodes")
fi
if [[ -z $node_host || $node_host == '<none>' ]]; then
  echo 'Helm succeeded, but no Ready node address was found. Set LAKEHOUSE_ACCESS_HOST and rerun this access check.' >&2
  exit 1
fi
if [[ ! $node_host =~ ^[A-Za-z0-9._:-]+$ ]]; then
  echo 'Helm succeeded, but the node address contains unsupported characters.' >&2
  exit 1
fi
url_host=$node_host
if [[ $url_host == *:* ]]; then url_host="[$url_host]"; fi

dashboard_service="oceanbase-dashboard-$release"
if kubectl -n "$namespace" get service "$dashboard_service" >/dev/null 2>&1; then
  dashboard_type=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.spec.type}')
  case $dashboard_type in
    NodePort)
      dashboard_port=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.spec.ports[0].nodePort}')
      if [[ -z $dashboard_port ]]; then echo 'Dashboard NodePort is not allocated.' >&2; exit 1; fi
      printf '\nDashboard URL: http://%s:%s/\n' "$url_host" "$dashboard_port"
      ;;
    LoadBalancer)
      dashboard_address=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
      if [[ -z $dashboard_address ]]; then
        dashboard_address=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
      fi
      dashboard_port=$(kubectl -n "$namespace" get service "$dashboard_service" -o jsonpath='{.spec.ports[0].port}')
      if [[ -n $dashboard_address ]]; then
        printf '\nDashboard URL: http://%s:%s/\n' "$dashboard_address" "$dashboard_port"
      else
        echo 'Dashboard LoadBalancer address is still pending.'
      fi
      ;;
    *) printf '\nDashboard is internal. Run: kubectl -n %s port-forward svc/%s 8080:80\n' "$namespace" "$dashboard_service" ;;
  esac
else
  echo 'Dashboard Service is absent; check dashboard.enabled in the release values.'
fi

sql_service="$release-sql"
if kubectl -n "$namespace" get service "$sql_service" >/dev/null 2>&1; then
  sql_type=$(kubectl -n "$namespace" get service "$sql_service" -o jsonpath='{.spec.type}')
  if [[ $sql_type == NodePort ]]; then
    sql_port=$(kubectl -n "$namespace" get service "$sql_service" -o jsonpath='{.spec.ports[0].nodePort}')
    if [[ -z $sql_port ]]; then echo 'SQL NodePort is not allocated.' >&2; exit 1; fi
    printf "SQL command: mysql -h '%s' -P %s -u 'root@sys' -p\n" "$node_host" "$sql_port"
  else
    printf 'SQL is internal. Run: kubectl -n %s port-forward svc/%s 12881:2881\n' "$namespace" "$sql_service"
  fi
else
  echo 'SQL Service is absent; check stage=cluster and the release status.'
fi
echo 'Use the password retrieval commands from Helm NOTES in a trusted terminal.'
