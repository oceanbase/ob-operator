#!/bin/bash
set -euo pipefail
source setup.sh
source util.sh
source case_p7_TenantParamManage/tenant_settings_fixture.sh

prepare_p7_fixture
start_p7_tenant
children() {
    kubectl -n "$NAMESPACE" get obtenantvariable -o json | jq -c --arg tenant "$OBTENANT_NAME" '
      [.items[] | select(.spec.obtenant == $tenant and .spec.variable.name == "autocommit")]'
}

for value in 1 0; do
    patch=$(jq -cn --arg value "$value" '{spec:{variables:[{name:"autocommit",value:$value}]}}')
    kubectl -n "$NAMESPACE" patch "obtenant/$OBTENANT_NAME" --type=merge -p "$patch"
    for ((attempt=0; attempt<100; attempt++)); do
        if jq -e --arg value "$value" 'length == 1 and .[0].status.status == "matched" and .[0].spec.variable.value == $value' <<< "$(children)" >/dev/null; then
            actual=$(tenant_sql 'SELECT @@GLOBAL.autocommit')
            [[ $actual == "$value" ]] && break
        fi
        sleep 3
    done
    ((attempt < 100))
    echo "$OBTENANT_NAME autocommit=$value in the tenant database and matched OBTenantVariable"
done
kubectl -n "$NAMESPACE" patch "obtenant/$OBTENANT_NAME" --type=json -p '[{"op":"remove","path":"/spec/variables"}]'
for ((attempt=0; attempt<100; attempt++)); do
    if [[ $(children) == '[]' ]]; then break; fi
    sleep 3
done
((attempt < 100))
echo "$OBTENANT_NAME autocommit OBTenantVariable removed"
echo "case passed"
