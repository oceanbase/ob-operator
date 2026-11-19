#!/usr/bin/env bash
set -euo pipefail

namespace=${OPERATOR_NS:-oceanbase-system}
deployment=${OPERATOR_DEPLOYMENT:-oceanbase-controller-manager}
env_name=OB_OPERATOR_TASK_POOLSIZE
k=(kubectl -n "$namespace")
original=$("${k[@]}" get deployment "$deployment" -o json |
    jq -ce --arg key "$env_name" '
      .spec.template.spec.containers[] | select(.name == "manager") |
      ([.env[]? | select(.name == $key)][0] // {name: $key, "$patch": "delete"})')

restore_config() {
    local rc=$? patch
    trap - EXIT
    # A previous valueFrom must replace the literal value installed by the test.
    patch=$(jq -nc --argjson entry "$original" '
      {spec: {template: {spec: {containers: [{name: "manager",
        env: [(if $entry.valueFrom then {value: null} else {} end) + $entry]}]}}}}')
    "${k[@]}" patch deployment "$deployment" --type=strategic -p "$patch" &&
        "${k[@]}" rollout status "deployment/$deployment" --timeout=3m || rc=1
    if (( rc != 0 )); then echo "case failed"; else echo "case passed"; fi
    exit "$rc"
}
trap restore_config EXIT

for value in 9999 10000; do
    "${k[@]}" set env "deployment/$deployment" --containers=manager "$env_name=$value"
    "${k[@]}" rollout status "deployment/$deployment" --timeout=3m
    # Completed rollouts may still have terminating Pods from the old revision.
    current=$("${k[@]}" get deployment "$deployment" -o json)
    replica_set=$("${k[@]}" get replicasets -o json | jq -er --argjson d "$current" '
      .items[] | select(.metadata.annotations["deployment.kubernetes.io/revision"] ==
        $d.metadata.annotations["deployment.kubernetes.io/revision"]) |
      select(any(.metadata.ownerReferences[]?; .uid == $d.metadata.uid)) | .metadata.uid')
    pods=$("${k[@]}" get pods -o json | jq -er --arg uid "$replica_set" '
      .items[] | select(.metadata.deletionTimestamp == null and
        any(.metadata.ownerReferences[]?; .uid == $uid)) | .metadata.name')
    while IFS= read -r pod; do
        "${k[@]}" logs "$pod" -c manager | jq -Rse --argjson value "$value" '
          any(split("\n")[] | fromjson?; .msg == "starting manager" and .configs.Task.PoolSize == $value)'
    done <<< "$pods"
    echo "manager started with Task.PoolSize=$value"
done
