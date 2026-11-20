#!/bin/bash
set -euo pipefail
source setup.sh
source util.sh

source case_p7_TenantParamManage/tenant_settings_fixture.sh
prepare_p7_fixture
start_p7_tenant

# Keep the vector test on releases offering it, and exercise a dynamic tenant parameter on older releases.
parameter=$(tenant_sql "SELECT DISTINCT NAME FROM GV\$OB_PARAMETERS WHERE NAME IN ('ob_vector_memory_limit_percentage', 'log_archive_concurrency') ORDER BY NAME DESC LIMIT 1")
case $parameter in
    ob_vector_memory_limit_percentage) values=(20 30) ;;
    log_archive_concurrency) values=(2 3) ;;
    *) echo "No supported dynamic tenant parameter" >&2; exit 1 ;;
esac
children() {
    kubectl -n "$NAMESPACE" get obparameter -o json | jq -c --arg tenant "$OBTENANT_NAME" --arg parameter "$parameter" '
      [.items[] | select(.spec.tenantName == $tenant and .spec.parameter.name == $parameter)]'
}
for value in "${values[@]}"; do
    patch=$(jq -cn --arg name "$parameter" --arg value "$value" '{spec:{parameters:[{name:$name,value:$value}]}}')
    kubectl -n "$NAMESPACE" patch "obtenant/$OBTENANT_NAME" --type=merge -p "$patch"
    for ((attempt=0; attempt<100; attempt++)); do
        if jq -e --arg value "$value" 'length == 1 and .[0].status.status == "matched" and .[0].spec.parameter.value == $value' <<< "$(children)" >/dev/null; then
            actual=$(tenant_sql "SELECT DISTINCT VALUE FROM GV\$OB_PARAMETERS WHERE NAME='$parameter'")
            [[ $actual == "$value" ]] && break
        fi
        sleep 3
    done
    ((attempt < 100))
    echo "$OBTENANT_NAME $parameter=$value in the tenant database and matched OBParameter"
done
kubectl -n "$NAMESPACE" patch "obtenant/$OBTENANT_NAME" --type=json -p '[{"op":"remove","path":"/spec/parameters"}]'
for ((attempt=0; attempt<100; attempt++)); do
    if [[ $(children) == '[]' ]]; then break; fi
    sleep 3
done
((attempt < 100))
echo "$OBTENANT_NAME $parameter OBParameter removed"
echo "case passed"
