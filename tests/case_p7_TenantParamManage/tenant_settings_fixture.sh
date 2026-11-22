# shellcheck shell=bash
# Shared fixture for the P7 tenant SQL and cluster display scenarios.
prepare_p7_fixture() {
    SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    PASSWORD=$(generate_random_str)
    export SUFFIX PASSWORD
    export NAMESPACE="$NAMESPACE_PREFIX-$SUFFIX" OBCLUSTER_NAME="test$SUFFIX"
    export OB_ROOT_SECRET="sc-root-$SUFFIX" OBTENANT_NAME="tenant$SUFFIX"
    kubectl create namespace "$NAMESPACE"
    trap cleanup_p7_fixture EXIT
    create_pass_secret "$NAMESPACE" "$OB_ROOT_SECRET" "$PASSWORD"

    local file key
    file=$(mktemp case_p7_TenantParamManage/env_vars.XXXXXX)
    for key in SUFFIX NAMESPACE OBCLUSTER_NAME OB_ROOT_SECRET OBTENANT_NAME; do
        printf 'export %s=%q\n' "$key" "${!key}"
    done > "$file"
    mv -- "$file" case_p7_TenantParamManage/env_vars.sh
}

cleanup_p7_fixture() {
    local rc=$?
    trap - EXIT
    kubectl delete namespace "$NAMESPACE" --timeout=5m || rc=1
    if ((rc != 0)); then echo "case failed" >&2; fi
    exit "$rc"
}

start_p7_tenant() {
    local zones secret
    envsubst < config/tenantManage/obcluster_template_1-1-1.yaml | kubectl apply -f -
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=8m
    p7_host=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath='{.items[0].status.podIP}')
    zones=$(MYSQL_PWD="$PASSWORD" mysql --batch --skip-column-names --connect-timeout=3 -h "$p7_host" -P2881 -uroot -Doceanbase -e "SELECT ZONE FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE' ORDER BY ZONE")
    [[ $zones == $'zone1\nzone2\nzone3' ]]
    envsubst < config/tenantManage/obtenant_basic_1_1_1.yaml | kubectl apply -f -
    kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obtenant/$OBTENANT_NAME" --timeout=8m
    p7_tenant_password=''
    secret=$(kubectl -n "$NAMESPACE" get "obtenant/$OBTENANT_NAME" -o jsonpath='{.spec.credentials.root}')
    if [[ -n $secret ]]; then
        p7_tenant_password=$(kubectl -n "$NAMESPACE" get "secret/$secret" -o jsonpath='{.data.password}' | base64 -d)
    fi
}

tenant_sql() {
    MYSQL_PWD="$p7_tenant_password" mysql --connect-timeout=3 --batch --skip-column-names -h "$p7_host" -P2881 -u "root@$OBTENANT_NAME" -Doceanbase -e "$1"
}
