#!/bin/bash

source setup.sh
source util.sh
source env.sh
source case_p3_AdvancedConfig/env_vars.sh

apply_and_check() {
    local template=$1 expected_cpu=$2 expected_memory=$3
    local attempt cluster_status pod ip cpu memory observer_ready agent_ready active
    local columns='{range .items[*]}{.status.podIP}{" "}{.spec.containers[?(@.name=="observer")].resources.requests.cpu}{" "}{.spec.containers[?(@.name=="observer")].resources.requests.memory}{" "}{.status.containerStatuses[?(@.name=="observer")].ready}{" "}{.status.containerStatuses[?(@.name=="obagent")].ready}{"\n"}{end}'

    envsubst < "./config/2.1.2_test/$template" | kubectl apply -f - || return 1
    for ((attempt=0; attempt<100; attempt++)); do
        cluster_status=$(kubectl -n "$NAMESPACE" get "obcluster/$OBCLUSTER_NAME" -o jsonpath='{.status.status}') || return 1
        pod=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath="$columns") || return 1
        read -r ip cpu memory observer_ready agent_ready <<< "$pod"
        if [[ $cluster_status == running && $cpu == "$expected_cpu" && $memory == "$expected_memory" && $observer_ready == true && $agent_ready == true ]]; then
            active=$(MYSQL_PWD="$PASSWORD" mysql --connect-timeout=3 -N -B -uroot -h "$ip" -P 2881 -Doceanbase \
                -e "SELECT COUNT(*) FROM DBA_OB_SERVERS WHERE ZONE='zone1' AND SVR_IP='127.0.0.1' AND STATUS='ACTIVE'") || active=0
            if [[ $active == 1 ]]; then
                echo "Observer ready with $cpu CPU, $memory memory and an ACTIVE standalone SQL server"
                return 0
            fi
        fi
        sleep 3
    done
    return 1
}

if apply_and_check cluster_standalone.yaml 2 10Gi && apply_and_check cluster_standalone_scaled_up.yaml 3 12Gi; then
    echo "case passed"
else
    echo "case failed"
    exit 1
fi
