#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p3_AdvancedConfig/env_vars.sh

validate() {
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=5m || return 1
    kubectl -n "$NAMESPACE" wait --for=condition=Ready pod -l "ref-obcluster=$OBCLUSTER_NAME" --timeout=5m || return 1
    local standbyIP status
    standbyIP=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME_STANDBY" -o jsonpath='{.items[0].status.podIP}') || return 1
    status=$(MYSQL_PWD="$PASSWORD" mysql -N -B -uroot -h "$standbyIP" -P 2881 -Doceanbase \
        -e "SELECT STATUS FROM DBA_OB_SERVERS WHERE ZONE = 'zone1'") || return 1
    [[ $status == ACTIVE ]] || return 1
    envsubst < ./config/2.1.2_test/primary_tenant.yaml | kubectl apply -f - || return 1
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obtenant/$TENANT_PRIMARY" --timeout=5m || return 1
    local credential
    credential=$(kubectl -n "$NAMESPACE" get "obtenant/$TENANT_PRIMARY" -o jsonpath='{.spec.credentials.standbyRo}') || return 1
    [[ $credential == "$TENANT_STANDBY_RO_SECRET" ]] || return 1
    kubectl -n "$NAMESPACE" get "secret/$credential" -o go-template='{{if .data.password}}password configured{{end}}' | grep -qx 'password configured'
}

if validate; then
    echo "case passed"
else
    echo "case failed"
    exit 1
fi
