#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p3_AdvancedConfig/env_vars.sh
# prepare create related resources to create obcluster
prepare() {
    export OBCLUSTER_NAME_STANDBY=obclusterstandby$SUFFIX
    export TENANT_PRIMARY=tenantprimary$SUFFIX
    export TENANT_STANDBY=tenantstandby$SUFFIX
    export TENANT_STANDBY_RO_SECRET=standby-ro-$SUFFIX
}
 
run() {
    envsubst < "./config/2.1.2_test/$1" | kubectl apply -f -
}

check_cluster_running() {
    local cluster=$1 ip state
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$cluster" --timeout=5m || return 1
    kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l "ref-obcluster=$cluster" --timeout=5m || return 1
    ip=$(kubectl -n "$NAMESPACE" get pod -l "ref-obcluster=$cluster" -o jsonpath='{.items[0].status.podIP}') || return 1
    state=$(MYSQL_PWD="$PASSWORD" mysql -N -B -uroot -h "$ip" -P 2881 -Doceanbase \
        -e "SELECT STATUS FROM DBA_OB_SERVERS WHERE ZONE = 'zone1'") || return 1
    [[ $state == ACTIVE ]]
}

wait_for_primary_pool_capacity() {
    local ip available deadline=$((SECONDS + 300))
    ip=$(kubectl -n "$NAMESPACE" get pod -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath='{.items[0].status.podIP}') || return 1
    while (( SECONDS < deadline )); do
        available=$(MYSQL_PWD="$PASSWORD" mysql -N -B -uroot -h "$ip" -P 2881 -Doceanbase \
            -e "SELECT MEM_CAPACITY - MEM_ASSIGNED >= 5 * 1024 * 1024 * 1024 FROM GV\$OB_SERVERS WHERE ZONE = 'zone1'") || return 1
        [[ $available == 1 ]] && return 0
        sleep 2
    done
    return 1
}

export_to_file() {
    local output_file="case_p3_AdvancedConfig/env_vars.sh"
    {
        echo "export OBCLUSTER_NAME_STANDBY=\"$OBCLUSTER_NAME_STANDBY\""
	echo "export TENANT_PRIMARY=\"$TENANT_PRIMARY\""
	echo "export TENANT_STANDBY=\"$TENANT_STANDBY\""
	echo "export TENANT_STANDBY_RO_SECRET=\"$TENANT_STANDBY_RO_SECRET\""
    } >> "$output_file"
    echo "Environment variables have been appended to $output_file"
}


validate() {
    # OceanBase 4.2 reserves 5Gi for sys. Leave room for the 5Gi primary tenant.
    envsubst < ./config/2.1.2_test/cluster_standalone.yaml |
        kubectl patch --local -f - --type=merge -p '{"spec":{"observer":{"resource":{"memory":"14Gi"}}}}' -o yaml |
        kubectl apply -f - || return 1
    run "cluster_standalone_standby.yaml" || return 1
    check_cluster_running "$OBCLUSTER_NAME" || return 1
    check_cluster_running "$OBCLUSTER_NAME_STANDBY" || return 1
    local tenant secret output
    tenant=$(kubectl -n "$NAMESPACE" get "obtenant/$TENANT_PRIMARY" --ignore-not-found -o name) || return 1
    if [[ -z $tenant ]]; then
        secret=$(kubectl -n "$NAMESPACE" get "secret/$TENANT_STANDBY_RO_SECRET" --ignore-not-found -o name) || return 1
        [[ -z $secret ]] || return 1
        wait_for_primary_pool_capacity || return 1
    fi
    run "primary_tenant.yaml" || return 1
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obtenant/$TENANT_PRIMARY" --timeout=5m || return 1
    if output=$(run "standby_tenant.yaml" 2>&1); then
        return 1
    fi
    grep -Fq "Given tenant is in a standalone cluster, which can not be a restore source" <<< "$output"
}

prepare
export_to_file
if validate; then
    echo "case passed"
else
    echo "case failed"
    exit 1
fi
