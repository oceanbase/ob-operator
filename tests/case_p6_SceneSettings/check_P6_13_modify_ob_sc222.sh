#!/bin/bash
set -euo pipefail
source setup.sh
source util.sh

SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
PASSWORD=$(generate_random_str)
export NAMESPACE="$NAMESPACE_PREFIX-$SUFFIX" OBCLUSTER_NAME="test$SUFFIX"
export OB_ROOT_SECRET="sc-root-$SUFFIX" LOCAL_PATH_NEW="local-path-new$SUFFIX" PASSWORD
templates=config/2.3.0_test
evidence=$(mktemp -d)
ns_created=false
sc_created=false
cleanup() {
    local rc=$?
    trap - EXIT
    if [[ $ns_created == true ]]; then
        if ((rc != 0)); then kubectl -n "$NAMESPACE" get obcluster,obzone,observer,pod,pvc -o wide >&2 || true; fi
        kubectl delete namespace "$NAMESPACE" --timeout=5m || rc=1
    fi
    if [[ $sc_created == true ]]; then kubectl delete storageclass "$LOCAL_PATH_NEW" || rc=1; fi
    rm -r "$evidence" || rc=1
    exit "$rc"
}
trap cleanup EXIT
kubectl create namespace "$NAMESPACE"
ns_created=true
create_pass_secret "$NAMESPACE" "$OB_ROOT_SECRET" "$PASSWORD"
kubectl -n "$NAMESPACE" create serviceaccount "$NAMESPACE"
envsubst < "$templates/new_sc.yaml" | kubectl create -f -
sc_created=true
cluster=$(envsubst < "$templates/obcluster_template_222.yaml")
topology=$(kubectl create --dry-run=client -f - -o json <<< "$cluster")
expected_zones=$(jq -r '.spec.topology[] | [.zone,.replica] | @tsv' <<< "$topology" | sort)
replicas=$(jq '[.spec.topology[].replica] | add' <<< "$topology")
prefix="$OBCLUSTER_NAME-$(jq -r '.spec.clusterId' <<< "$topology")-"

# Each row identifies one Ready Observer and the claims that Pod actually mounts.
pod_snapshot() {
    kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o json > "$evidence/pod-list.json" || return 1
    jq -r --arg prefix "$prefix" '
      .items[] | select(.metadata.deletionTimestamp == null and
        any(.status.containerStatuses[]?; .name == "observer" and .ready)) |
      [(.metadata.labels["ref-obzone"] | ltrimstr($prefix)), .metadata.name, .metadata.uid] +
      [.spec.volumes[] | select(.persistentVolumeClaim) | .persistentVolumeClaim.claimName] | @tsv' "$evidence/pod-list.json" | sort
}

placement_matches() {
    local actual
    actual=$(cut -f1 "$1" | sort | uniq -c | awk '{print $2 "\t" $1}')
    [[ $actual == "$expected_zones" ]]
}

# Resolve each mounted claim to the matching Bound PV, including the claim UID.
bound_volumes() {
    local snapshot=$1 class=$2
    local -a claims volumes
    mapfile -t claims < <(awk -F '\t' '{for (i=4; i<=NF; i++) print $i}' "$snapshot" | sort -u)
    ((${#claims[@]} == 3 * replicas)) || return 1
    kubectl -n "$NAMESPACE" get pvc "${claims[@]}" -o json > "$evidence/claims.json" || return 1
    mapfile -t volumes < <(jq -r '.items[] | .spec.volumeName // empty' "$evidence/claims.json" | sort -u)
    ((${#volumes[@]} == ${#claims[@]})) || return 1
    kubectl get pv "${volumes[@]}" -o json > "$evidence/volumes.json" || return 1
    jq -e --arg ns "$NAMESPACE" --arg sc "$class" --slurpfile claims "$evidence/claims.json" '
      . as $vs | all($claims[0].items[]; . as $c |
        .status.phase == "Bound" and .spec.storageClassName == $sc and
        any($vs.items[]; .metadata.name == $c.spec.volumeName and .status.phase == "Bound" and
          .spec.storageClassName == $sc and .spec.claimRef.uid == $c.metadata.uid and
          .spec.claimRef.namespace == $ns and .spec.claimRef.name == $c.metadata.name))' "$evidence/volumes.json" >/dev/null || return 1
    jq -r '.items[] | [.metadata.uid,.metadata.name,.spec.storageClassName,.spec.claimRef.uid,.spec.claimRef.name] | @tsv' "$evidence/volumes.json" | sort
}

sys_sql() {
    local snapshot=$1 query=$2 pod ip
    pod=$(awk -F '\t' 'NR==1 {print $2}' "$snapshot")
    ip=$(kubectl -n "$NAMESPACE" get pod "$pod" -o jsonpath='{.status.podIP}') || return 1
    MYSQL_PWD="$PASSWORD" mysql --connect-timeout=3 --batch --skip-column-names -h "$ip" -P2881 -uroot -Doceanbase -e "$query"
}

kubectl apply -f - <<< "$cluster"
kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=10m
pod_snapshot > "$evidence/old-pods"
placement_matches "$evidence/old-pods"
bound_volumes "$evidence/old-pods" local-path > "$evidence/old-volumes"
sys_sql "$evidence/old-pods" "CREATE DATABASE storage_migration; CREATE TABLE storage_migration.proof (id INT PRIMARY KEY); INSERT INTO storage_migration.proof VALUES(13)"
envsubst < "$templates/obcluster_template_new_sc222.yaml" | kubectl apply -f -

migration_timeout=$((300 * (replicas + 1)))
deadline=$((SECONDS + migration_timeout))
peak=$replicas
while ((SECONDS < deadline)); do
    if ! pod_snapshot > "$evidence/new-pods"; then sleep 3; continue; fi
    servers=$(kubectl -n "$NAMESPACE" get observers -l "ref-obcluster=$OBCLUSTER_NAME" -o json) || { sleep 3; continue; }
    server_count=$(jq '.items | length' <<< "$servers")
    if ((server_count > replicas + 1)); then
        echo "case failed: $server_count Observer resources exceed the rollout budget of $((replicas + 1))" >&2
        exit 1
    fi
    scheduled=$(jq '[.items[] | select(.metadata.deletionTimestamp == null and .spec.nodeName != null and
      any(.spec.containers[]; .name == "observer"))] | length' "$evidence/pod-list.json")
    if ((scheduled > peak)); then
        peak=$scheduled
        echo "Observed $peak scheduled Observer Pods during migration"
    fi
    if placement_matches "$evidence/new-pods" &&
        [[ -z $(comm -12 <(cut -f3 "$evidence/old-pods" | sort) <(cut -f3 "$evidence/new-pods" | sort)) ]] &&
        [[ $(kubectl -n "$NAMESPACE" get "obcluster/$OBCLUSTER_NAME" -o jsonpath='{.status.status}') == running ]] &&
        bound_volumes "$evidence/new-pods" "$LOCAL_PATH_NEW" > "$evidence/new-volumes" &&
        [[ -z $(comm -12 <(cut -f1 "$evidence/old-volumes" | sort) <(cut -f1 "$evidence/new-volumes" | sort)) ]] &&
        [[ -z $(comm -12 <(cut -f4 "$evidence/old-volumes" | sort) <(cut -f4 "$evidence/new-volumes" | sort)) ]] &&
        zones=$(sys_sql "$evidence/new-pods" "SELECT ZONE, COUNT(*) FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE' GROUP BY ZONE ORDER BY ZONE") &&
        [[ $zones == "$expected_zones" ]] &&
        [[ $(sys_sql "$evidence/new-pods" 'SELECT COUNT(*) FROM storage_migration.proof WHERE id=13') == 1 ]]; then
        for phase in old new; do
            echo "$phase Observer and volume identities"
            cat "$evidence/$phase-pods" "$evidence/$phase-volumes"
        done
        echo "Peak sampled scheduled Observer Pods $peak"
        echo "Verified $replicas ACTIVE Observers, $((3*replicas)) replacement PVCs and PVs, and original SQL row 13"
        echo "case passed"
        exit 0
    fi
    sleep 3
done
echo "case failed: storage migration did not replace the original mounts and restore SQL within ${migration_timeout}s" >&2
exit 1
