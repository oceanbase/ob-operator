#!/bin/bash
set -euo pipefail
source setup.sh
source util.sh
source env.sh
source case_p0_ClusterAndTenant/env_vars.sh

: "${NAMESPACE:?}" "${OBCLUSTER_NAME:?}" "${PASSWORD:?}" "${OB_ROOT_SECRET:?}"
[[ $NAMESPACE == "$NAMESPACE_PREFIX-${SUFFIX:?}" && $OBTENANT_NAME =~ ^[[:alnum:]]+$ ]]
reset_namespace() {
    kubectl delete namespace "$NAMESPACE" --ignore-not-found
    kubectl create namespace "$NAMESPACE"
    create_pass_secret "$NAMESPACE" "$OB_ROOT_SECRET" "$PASSWORD"
}
apply_template() {
    envsubst < "config/tenantManage/$1" | kubectl apply -f -
}
sql() {
    MYSQL_PWD="$PASSWORD" mysql --connect-timeout=3 --batch --skip-column-names \
        -h "$ip" -P2881 -uroot@sys -Doceanbase -e "$1"
}
wait_units() {
    local count=$1 deadline=$((SECONDS + 720)) state='' actual='' expected
    printf -v expected 'zone1\t%s\t%s\t%s\nzone2\t%s\t%s\t%s\nzone3\t%s\t%s\t%s' \
        "$count" "$count" "$count" "$count" "$count" "$count" "$count" "$count" "$count"
    while ((SECONDS < deadline)); do
        state=$(kubectl -n "$NAMESPACE" get "obtenant/$OBTENANT_NAME" -o jsonpath='{.status.status}') || state=unavailable
        actual=$(sql "SELECT P.ZONE_LIST,P.UNIT_COUNT,COUNT(U.UNIT_ID),COUNT(DISTINCT U.SVR_IP)
            FROM DBA_OB_TENANTS T JOIN DBA_OB_RESOURCE_POOLS P ON T.TENANT_ID=P.TENANT_ID
            JOIN DBA_OB_UNITS U ON U.RESOURCE_POOL_ID=P.RESOURCE_POOL_ID
            WHERE T.TENANT_NAME='$OBTENANT_NAME' AND T.STATUS='NORMAL' AND U.STATUS='ACTIVE'
            GROUP BY P.ZONE_LIST,P.UNIT_COUNT ORDER BY P.ZONE_LIST") || actual=unavailable
        if [[ $state == running && $actual == "$expected" ]]; then
            printf 'Tenant %s has %s active units on distinct servers per zone\n%s\n' "$OBTENANT_NAME" "$count" "$actual"
            return 0
        fi
        [[ ${state,,} != failed ]] || break
        sleep 3
    done
    printf 'case failed waiting for %s units per zone (tenant state %s)\n%s\n' "$count" "$state" "$actual" >&2
    return 1
}
tenant_sql() {
    local host
    host=$(sql "SELECT U.SVR_IP FROM DBA_OB_UNITS U JOIN DBA_OB_RESOURCE_POOLS P USING (RESOURCE_POOL_ID)
        JOIN DBA_OB_TENANTS T USING (TENANT_ID) WHERE T.TENANT_NAME='$OBTENANT_NAME' AND U.STATUS='ACTIVE'
        ORDER BY U.ZONE,U.UNIT_ID LIMIT 1")
    [[ -n $host ]]
    MYSQL_PWD= mysql --connect-timeout=3 --batch --skip-column-names -h "$host" -P2881 -u "root@$OBTENANT_NAME" -e "$1"
}

reset_namespace
apply_template obcluster_template_2-2-2.yaml
kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=15m
ip=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME,ref-observer" -o json |
    jq -er '[.items[] | select(.metadata.deletionTimestamp == null and .status.phase == "Running")][0].status.podIP')
servers=$(sql "SELECT ZONE,COUNT(*) FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE' GROUP BY ZONE ORDER BY ZONE")
[[ $servers == $'zone1\t2\nzone2\t2\nzone3\t2' ]]
printf 'Active servers by zone\n%s\n' "$servers"
apply_template obtenant_basic_1_1_1.yaml
wait_units 1
tenant_sql "CREATE DATABASE unit_count_test; CREATE TABLE unit_count_test.proof(id INT PRIMARY KEY, marker VARCHAR(80));
    INSERT INTO unit_count_test.proof VALUES (1,'$OBTENANT_NAME')"
apply_template obtenant_basic_2_2_2.yaml
wait_units 2
[[ $(tenant_sql "SELECT marker FROM unit_count_test.proof WHERE id=1") == "$OBTENANT_NAME" ]]
echo "Tenant SQL row retained after unit-count expansion"
reset_namespace
echo "case passed"
