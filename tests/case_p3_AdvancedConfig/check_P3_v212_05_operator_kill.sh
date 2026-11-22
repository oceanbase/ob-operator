#!/usr/bin/env bash
set -o pipefail

source setup.sh
source case_p3_AdvancedConfig/env_vars.sh

operator_ns=${OPERATOR_NS:-oceanbase-system}
export OBCLUSTER_NAME_THREE_NODE="testthreenode$SUFFIX"

sys_sql() {
    local ip
    ip=$(kubectl -n "$NAMESPACE" get pods -l "ref-obzone=$OBCLUSTER_NAME_THREE_NODE-1-zone2" -o jsonpath='{.items[0].status.podIP}') || return 1
    MYSQL_PWD="$PASSWORD" mysql -N -B --connect-timeout=3 -uroot -h "$ip" -P 2881 -Doceanbase -e "$1"
}

# Select the elected manager, or the sole Ready manager on installations
# without leader election. A terminating previous Pod may still be listed.
leader_pod() {
    local holder
    holder=$(kubectl -n "$operator_ns" get lease operator.oceanbase.com --ignore-not-found \
        -o jsonpath='{.spec.holderIdentity}') || return 1
    kubectl -n "$operator_ns" get pods -l control-plane=controller-manager -o json |
        jq -er --arg leader "${holder%%_*}" '
          [.items[] | select(.metadata.deletionTimestamp == null and
             any(.status.conditions[]?; .type == "Ready" and .status == "True")) |
             select($leader == "" or .metadata.name == $leader)] |
          select(length == 1) | .[0] | [.metadata.name, .metadata.uid] | @tsv'
}

validate() {
    envsubst < ./config/2.1.2_test/cluster_three_nodes.yaml | kubectl apply -f - || return 1
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME_THREE_NODE" --timeout=8m || return 1
    local database_pod database_uid old_manager manager_uid new_manager_uid recovering new_pod recovered attempt leader
    database_pod=$(kubectl -n "$NAMESPACE" get pods -l "ref-obzone=$OBCLUSTER_NAME_THREE_NODE-1-zone1" -o jsonpath='{.items[0].metadata.name}') || return 1
    database_uid=$(kubectl -n "$NAMESPACE" get pod "$database_pod" -o jsonpath='{.metadata.uid}') || return 1
    leader=$(leader_pod) || return 1
    IFS=$'\t' read -r old_manager manager_uid <<< "$leader"
    local active="SELECT COUNT(DISTINCT ZONE) FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE'"
    [[ $(sys_sql "$active") == 3 ]] || return 1
    sys_sql "CREATE DATABASE IF NOT EXISTS test; CREATE TABLE IF NOT EXISTS test.p3_recovery (id INT PRIMARY KEY); REPLACE INTO test.p3_recovery VALUES(1)" || return 1

    kubectl -n "$NAMESPACE" delete pod "$database_pod" --wait=false || return 1
    for ((attempt=0; attempt<300; attempt++)); do
        if ! recovering=$(kubectl -n "$NAMESPACE" get "observer/$database_pod" --ignore-not-found -o jsonpath='{.status.status}'); then
            sleep 1
            continue
        fi
        if [[ $recovering == "recover observer" || $recovering == "observer unrecoverable" || $recovering == deleting ]]; then
            break
        fi
        if [[ -z $recovering ]] && kubectl -n "$NAMESPACE" get observers -l "ref-obzone=$OBCLUSTER_NAME_THREE_NODE-1-zone1" -o json |
            jq -e 'all(.items[]; .status.status != "running")' >/dev/null; then
            break
        fi
        sleep 1
    done
    ((attempt < 300)) || return 1
    echo "Zone1 recovery started (${recovering:-old Observer removed})"
    kubectl -n "$operator_ns" delete pod "$old_manager" --wait=false || return 1
    for ((attempt=0; attempt<150; attempt++)); do
        if leader=$(leader_pod); then
            new_manager_uid=${leader#*$'\t'}
            [[ -n $new_manager_uid && $new_manager_uid != "$manager_uid" ]] && break
        fi
        sleep 2
    done
    ((attempt < 150)) || return 1
    for ((attempt=0; attempt<360; attempt++)); do
        new_pod=$(kubectl -n "$NAMESPACE" get pods -l "ref-obzone=$OBCLUSTER_NAME_THREE_NODE-1-zone1" -o go-template='{{range .items}}{{if not .metadata.deletionTimestamp}}{{.metadata.name}}{{"\n"}}{{end}}{{end}}' | head -1)
        if [[ -n $new_pod ]]; then
            recovered=$(kubectl -n "$NAMESPACE" get pod "$new_pod" --ignore-not-found -o jsonpath='{.metadata.uid}') || return 1
            [[ -n $recovered && $recovered != "$database_uid" ]] && break
        fi
        sleep 2
    done
    ((attempt < 360)) || return 1
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "observer/$new_pod" --timeout=12m || return 1
    kubectl -n "$NAMESPACE" wait --for=condition=Ready "pod/$new_pod" --timeout=5m || return 1
    for ((attempt=0; attempt<60; attempt++)); do
        if [[ $(sys_sql "$active") == 3 && $(sys_sql "SELECT COUNT(*) FROM test.p3_recovery WHERE id=1") == 1 ]]; then
            break
        fi
        sleep 3
    done
    ((attempt < 60)) || return 1
    echo "Observer $database_uid -> $recovered; manager $manager_uid -> $new_manager_uid; three ACTIVE zones, SQL row recovered"
}

if validate; then
    echo "case passed"
else
    echo "case failed"
    exit 1
fi
