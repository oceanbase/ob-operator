#!/usr/bin/env bash
set -euo pipefail

mode=$1
cd -- "$(dirname -- "$0")/.."
source setup.sh
source case_p0_ClusterAndTenant/env_vars.sh
: "${NAMESPACE:?}" "${OBCLUSTER_NAME:?}" "${PASSWORD:?}"
kubectl_ns=(kubectl -n "$NAMESPACE")
envsubst < config/clusterManage/obcluster_template_1-1-1.yaml | kubectl apply -f -
cluster_id=$("${kubectl_ns[@]}" get obcluster "$OBCLUSTER_NAME" -o jsonpath='{.spec.clusterId}')
zone_prefix="$OBCLUSTER_NAME-$cluster_id-zone"

# Return the zone1 Pod and container identities only after all three zones are
# ready. Terminating Pods can coexist with their replacements during recovery.
ready_zone1() {
    "${kubectl_ns[@]}" get pods -l "ref-obcluster=$OBCLUSTER_NAME,ref-observer" -o json |
        jq -er --arg prefix "$zone_prefix" '
          [.items[] | select(.metadata.deletionTimestamp == null)] |
          select((map(.metadata.labels["ref-obzone"]) | sort) ==
                 [($prefix+"1"), ($prefix+"2"), ($prefix+"3")]) |
          select(all(.[]; any(.status.conditions[]?; .type == "Ready" and .status == "True"))) |
          .[] | select(.metadata.labels["ref-obzone"] == $prefix+"1") |
          [.metadata.name, .metadata.uid,
           (.status.containerStatuses[] | select(.name == "observer").containerID),
           .status.podIP] | select(all(.[]; type == "string" and length > 0)) | @tsv'
}

wait_for_cluster() {
    local old_uid=${1:-} old_container=${2:-} deadline=$((SECONDS + 300))
    local row current servers
    while (( SECONDS < deadline )); do
        if row=$(ready_zone1); then
            IFS=$'\t' read -r pod uid container ip <<< "$row"
            if [[ $mode == pod && $uid == "$old_uid" ]] ||
               [[ $mode == container && $container == "$old_container" ]]; then
                sleep 3
                continue
            fi
            if current=$("${kubectl_ns[@]}" get obcluster "$OBCLUSTER_NAME" -o jsonpath='{.status.status}') &&
                [[ $current == running ]] && servers=$(MYSQL_PWD="$PASSWORD" timeout 8s mysql \
                --connect-timeout=3 -h "$ip" -P 2881 -uroot -Doceanbase --batch --skip-column-names \
                -e 'SELECT ZONE, STATUS FROM DBA_OB_SERVERS ORDER BY ZONE, SVR_IP') &&
                [[ $servers == $'zone1\tACTIVE\nzone2\tACTIVE\nzone3\tACTIVE' ]]; then
                echo "zone1 pod=$pod uid=$uid container=$container; three zones ACTIVE"
                return 0
            fi
        fi
        sleep 3
    done
    echo "case failed: cluster $NAMESPACE/$OBCLUSTER_NAME did not reach the expected $mode recovery state" >&2
    return 1
}

wait_for_cluster
case $mode in
    container)
        # Inject the original failure into the database process inside this container.
        # A disconnect is possible when the container exits during kubectl exec.
        rc=0
        "${kubectl_ns[@]}" exec "$pod" -c observer -- pkill -KILL -x observer || rc=$?
        if [[ $rc != 0 && $rc != 137 && $rc != 143 ]]; then exit "$rc"; fi
        ;;
    pod)
        "${kubectl_ns[@]}" delete pod "$pod" --wait=false
        ;;
    *) exit 2 ;;
esac
wait_for_cluster "$uid" "$container"
echo "case passed"
