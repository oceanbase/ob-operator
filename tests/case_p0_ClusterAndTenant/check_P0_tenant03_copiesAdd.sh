#!/bin/bash
set -euo pipefail
source setup.sh
source case_p0_ClusterAndTenant/env_vars.sh

[[ $OBTENANT_NAME =~ ^[[:alnum:]]+$ ]]
apply_template() {
    envsubst < "config/tenantManage/$1" | kubectl apply -f -
}
apply_template obcluster_template_1-1-1.yaml
kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=8m
ip=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath='{.items[0].status.podIP}')
sql() {
    MYSQL_PWD="$PASSWORD" mysql --batch --skip-column-names --connect-timeout=3 -h "$ip" -P2881 -uroot -Doceanbase -e "$1"
}
[[ $(sql "SELECT ZONE FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE' ORDER BY ZONE") == $'zone1\nzone2\nzone3' ]]

apply_and_wait() {
    local template=$1 expected=$2 state job rows prior until
    local jobs="FROM DBA_OB_TENANT_JOBS WHERE JOB_TYPE='ALTER_TENANT_LOCALITY' AND TENANT_ID=(SELECT TENANT_ID FROM DBA_OB_TENANTS WHERE TENANT_NAME='$OBTENANT_NAME')"
    prior=$(sql "SELECT COALESCE(MAX(JOB_ID), 0) $jobs")
    apply_template "$template"
    until=$((SECONDS + 600))
    while ((SECONDS < until)); do
        state=$(kubectl -n "$NAMESPACE" get "obtenant/$OBTENANT_NAME" -o jsonpath='{.status.status}')
        job=$(sql "SELECT JOB_ID, JOB_STATUS, RESULT_CODE $jobs AND JOB_ID>$prior ORDER BY JOB_ID DESC LIMIT 1")
        if [[ $state == failed || $job == *$'\tFAILED\t'* || $job == *$'\tCANCELED\t'* ]]; then
            echo "Tenant locality failed for $expected (CR $state, SQL job $job)" >&2
            return 1
        fi
        rows=$(sql "SELECT T.STATUS, T.PREVIOUS_LOCALITY IS NULL, GROUP_CONCAT(P.ZONE_LIST ORDER BY P.ZONE_LIST)
          FROM DBA_OB_TENANTS T JOIN DBA_OB_RESOURCE_POOLS P ON T.TENANT_ID=P.TENANT_ID
          WHERE T.TENANT_NAME='$OBTENANT_NAME' GROUP BY T.TENANT_ID,T.STATUS,T.PREVIOUS_LOCALITY")
        if [[ $state == running && $rows == $'NORMAL\t1\t'"$expected" && ( -z $job || $job == *$'\tSUCCESS\t0' ) ]]; then
            echo "$OBTENANT_NAME pools $expected, CR $state, SQL job ${job:-already converged}"
            return 0
        fi
        sleep 3
    done
    echo "Timed out waiting for $expected (CR $state, SQL pools $rows, SQL job $job)" >&2
    return 1
}
apply_and_wait obtenant_basic_1_1_1.yaml zone1,zone2,zone3
apply_and_wait obtenant_basic_1_1.yaml zone1,zone2
apply_and_wait obtenant_basic_1_1_1.yaml zone1,zone2,zone3
echo "case passed"
