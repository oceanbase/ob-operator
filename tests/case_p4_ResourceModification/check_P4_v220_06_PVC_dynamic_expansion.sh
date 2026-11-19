#!/bin/bash

source setup.sh
source util.sh
source env.sh
source case_p4_ResourceModification/env_vars.sh

set -euo pipefail
export OBCLUSTER_NAME=testexpandable$SUFFIX
export LOCAL_PATH_EXPANDABLE=${P4_EXPANDABLE_STORAGE_CLASS:-local-path-expandable$SUFFIX}
owned_class=false
cluster_created=false

cleanup() {
    if $cluster_created; then
        kubectl -n "$NAMESPACE" delete obcluster "$OBCLUSTER_NAME" --ignore-not-found
    fi
    if $owned_class; then
        kubectl delete storageclass "$LOCAL_PATH_EXPANDABLE" --ignore-not-found
    fi
}

template() {
    envsubst < "./config/2.2.0_test/$1"
}

volume_bytes() {
    kubectl exec -n "$NAMESPACE" "$server" -- df -PB1 /home/admin/data-file | awk 'NR == 2 {print $2}'
}

sql() {
    MYSQL_PWD="$PASSWORD" mysql --batch --skip-column-names -h "$ip" -P2881 -uroot -e "$1"
}

trap cleanup EXIT
if [[ -z ${P4_EXPANDABLE_STORAGE_CLASS:-} ]]; then
    template sc_local_expandable.yaml | kubectl create -f -
    owned_class=true
fi
if [[ $(kubectl get sc "$LOCAL_PATH_EXPANDABLE" -o jsonpath='{.allowVolumeExpansion}') != true ]]; then
    echo "case failed: storage class $LOCAL_PATH_EXPANDABLE does not allow expansion" >&2
    exit 1
fi
template obcluster_expandable_1-1-1.yaml | kubectl create -f -
cluster_created=true
kubectl -n "$NAMESPACE" wait "obcluster/$OBCLUSTER_NAME" --for=jsonpath='{.status.status}'=running --timeout=12m
server=$(kubectl -n "$NAMESPACE" get observers -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath='{.items[0].metadata.name}')
kubectl -n "$NAMESPACE" wait "pod/$server" --for=condition=Ready --timeout=3m
ip=$(kubectl -n "$NAMESPACE" get pod "$server" -o jsonpath='{.status.podIP}')
[[ $(sql "SELECT COUNT(*) FROM oceanbase.__all_server WHERE zone='zone1' AND status='ACTIVE'") == 1 ]]
sql 'CREATE DATABASE p4_resize; CREATE TABLE p4_resize.marker (id INT PRIMARY KEY); INSERT INTO p4_resize.marker VALUES (61)'
before=$(volume_bytes)
pvc="$server-data-file"
[[ $(kubectl -n "$NAMESPACE" get pvc "$pvc" -o jsonpath='{.spec.resources.requests.storage}') == 60Gi ]]

template obcluster_expanded_1-1-1.yaml | kubectl apply -f -
required=$(numfmt --from=iec-i 61Gi)
deadline=$((SECONDS + 420))
while (( SECONDS < deadline )); do
    read -r phase request capacity < <(kubectl -n "$NAMESPACE" get pvc "$pvc" -o jsonpath='{.status.phase} {.spec.resources.requests.storage} {.status.capacity.storage}{"\n"}')
    if [[ $phase == Bound && $request == 61Gi && -n $capacity ]] && (( $(numfmt --from=auto "$capacity") >= required )); then
        after=$(volume_bytes)
        if (( after > before )); then
            kubectl -n "$NAMESPACE" wait "observer/$server" "obzone/$OBCLUSTER_NAME-1-zone1" "obcluster/$OBCLUSTER_NAME" --for=jsonpath='{.status.status}'=running --timeout=5m
            [[ $(sql 'SELECT id FROM p4_resize.marker') == 61 ]]
            echo "case passed: $pvc request=$request capacity=$capacity filesystem=$before->$after bytes, SQL marker retained"
            exit 0
        fi
    fi
    sleep 3
done
echo "case failed: $pvc request=$request capacity=$capacity filesystem=$before->$(volume_bytes) bytes; use a resize-capable P4_EXPANDABLE_STORAGE_CLASS" >&2
exit 1
