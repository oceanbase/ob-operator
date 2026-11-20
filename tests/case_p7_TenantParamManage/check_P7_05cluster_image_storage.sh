#!/bin/bash
set -euo pipefail
source setup.sh
source util.sh

source case_p7_TenantParamManage/tenant_settings_fixture.sh
prepare_p7_fixture
manifest=$(envsubst < config/tenantManage/obcluster_template_1-1-1.yaml)
expected=$(kubectl create --dry-run=client -f - -o json <<< "$manifest")
kubectl apply -f - <<< "$manifest"
kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=8m

# Discover column positions from the actual server-provided table, including its wide columns.
table=$(kubectl -n "$NAMESPACE" get "obcluster/$OBCLUSTER_NAME" -o wide)
printf '%s\n' "$table"
column() {
    awk -v name="$1" 'NR==1 {for(i=1;i<=NF;i++) if($i==name) col=i} NR==2 && col {print $col}' <<< "$table"
}
image=$(column IMAGE)
storage=$(column STORAGE | jq -cS .)
[[ $image == "$(jq -r '.spec.observer.image' <<< "$expected")" ]]
[[ $storage == "$(jq -cS '.spec.observer.storage' <<< "$expected")" ]]

pods=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o json)
jq -e --arg image "$image" '
  (.items | length) == 3 and all(.items[];
    any(.spec.containers[]; .name == "observer" and .image == $image) and
    any(.status.containerStatuses[]; .name == "observer" and .ready and (.imageID | length > 0)))
' <<< "$pods" >/dev/null
claims=$(kubectl -n "$NAMESPACE" get pvc -o json)
while IFS= read -r claim; do
    case $claim in
        *-data-file) kind=dataStorage ;;
        *-data-log) kind=redoLogStorage ;;
        *-log) kind=logStorage ;;
        *) echo "Unexpected Observer claim $claim" >&2; exit 1 ;;
    esac
    jq -e --arg claim "$claim" --arg kind "$kind" --argjson storage "$storage" '
      any(.items[]; .metadata.name == $claim and .status.phase == "Bound"
        and .spec.storageClassName == $storage[$kind].storageClass
        and .spec.resources.requests.storage == $storage[$kind].size)
    ' <<< "$claims" >/dev/null
done < <(jq -r '.items[].spec.volumes[] | select(.persistentVolumeClaim) | .persistentVolumeClaim.claimName' <<< "$pods")
[[ $(jq '[.items[].spec.volumes[] | select(.persistentVolumeClaim)] | length' <<< "$pods") == 9 ]]
ip=$(jq -r '.items[0].status.podIP' <<< "$pods")
zones=$(MYSQL_PWD="$PASSWORD" mysql --connect-timeout=3 --batch --skip-column-names -h "$ip" -P2881 -uroot -Doceanbase -e "SELECT ZONE FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE' ORDER BY ZONE")
[[ $zones == $'zone1\nzone2\nzone3' ]]
echo "Verified $image in three Ready Observer Pods, nine bound PVCs, and three ACTIVE SQL zones"
echo "case passed"
