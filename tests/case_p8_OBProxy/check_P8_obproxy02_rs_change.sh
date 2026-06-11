#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TESTS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$TESTS_DIR/setup.sh"
source "$TESTS_DIR/util.sh"
source "$TESTS_DIR/env.sh"
source "$TESTS_DIR/case_p8_OBProxy/env_vars.sh"

# Define labels after sourcing env files to avoid being overwritten
OBPROXY_LABEL="obproxy.oceanbase.com/obproxy"
REF_CLUSTER_LABEL="ref-obcluster"

get_obproxy_deployment_name() {
    kubectl get deployment -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null
}

get_rs_list() {
    local pod
    pod=$(kubectl get pod -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -n "$pod" ]]; then
        kubectl exec -n "$NAMESPACE" "pod/$pod" -- env 2>/dev/null | grep "^RS_LIST=" | cut -d'=' -f2-
    fi
}

verify_all_obproxy_pods_rs_list() {
    local expected="$1"
    local ok=0
    while read -r pod; do
        [[ -z "$pod" ]] && continue
        local phase
        phase=$(kubectl get pod -n "$NAMESPACE" "$pod" -o jsonpath='{.status.phase}' 2>/dev/null)
        [[ "$phase" != "Running" ]] && continue
        local v
        v=$(kubectl exec -n "$NAMESPACE" "pod/$pod" -- env 2>/dev/null | grep "^RS_LIST=" | cut -d'=' -f2-)
        if [[ "$v" != "$expected" ]]; then
            echo "  RS_LIST mismatch on pod $pod"
            ok=1
        fi
    done < <(kubectl get pod -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null)
    return $ok
}

get_deployment_generation() {
    kubectl get deployment -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" -o jsonpath='{.items[0].metadata.generation}' 2>/dev/null
}

# Debug: Get OBProxy status
get_obproxy_status() {
    kubectl get obproxy -n "$NAMESPACE" "$OBPROXY_NAME" -o jsonpath='{.status.status}{"\t"}{.status.operationContext}' 2>/dev/null
}

# Debug: Get controller logs for obproxy
get_obproxy_controller_logs() {
    local since="${1:-2m}"
    kubectl logs -n oceanbase-system -l control-plane=controller-manager -c manager --since="$since" 2>/dev/null | grep -i "obproxy" | tail -20
}

MONITOR_PID=""
MONITOR_LOG=""

# Sample OBProxy conditions and Deployment generation every 2s into a temp file.
# Run in background; call stop_condition_monitor to terminate.
start_condition_monitor() {
    MONITOR_LOG=$(mktemp /tmp/obproxy_monitor.XXXXXX)
    (
        while true; do
            local ts gen rs_avail cluster_ready
            ts=$(date +%T)
            gen=$(kubectl get deployment -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" \
                -o jsonpath='{.items[0].metadata.generation}' 2>/dev/null)
            rs_avail=$(kubectl get obproxy "$OBPROXY_NAME" -n "$NAMESPACE" \
                -o jsonpath='{.status.conditions[?(@.type=="RSListAvailable")].reason}' 2>/dev/null)
            cluster_ready=$(kubectl get obproxy "$OBPROXY_NAME" -n "$NAMESPACE" \
                -o jsonpath='{.status.conditions[?(@.type=="OBClusterReady")].reason}' 2>/dev/null)
            echo "$ts gen=$gen RSListAvail=$rs_avail ClusterReady=$cluster_ready" >> "$MONITOR_LOG"
            sleep 2
        done
    ) &
    MONITOR_PID=$!
}

stop_condition_monitor() {
    if [[ -n "${MONITOR_PID:-}" ]]; then
        kill "$MONITOR_PID" 2>/dev/null
        wait "$MONITOR_PID" 2>/dev/null
        MONITOR_PID=""
    fi
}

# Assert debounce correctness from the monitor log.
# $1: gen_before  $2: label
analyze_debounce() {
    local gen_before="$1" label="$2"
    echo ""
    echo "=== Debounce Analysis: $label ==="
    if [[ ! -f "$MONITOR_LOG" ]]; then
        echo "  No monitor data."
        return
    fi

    echo "Condition timeline (sampled every 2s):"
    cat "$MONITOR_LOG"
    echo ""

    # Guard trigger: was OBClusterNotStable seen at least once?
    local not_stable_count
    not_stable_count=$(grep -c "OBClusterNotStable" "$MONITOR_LOG" 2>/dev/null || echo 0)
    if [[ "$not_stable_count" -gt 0 ]]; then
        echo "PASS (guard): OBClusterNotStable observed ${not_stable_count}x — controller skipped drift check while OBCluster was operating"
    else
        echo "NOTE (guard): OBClusterNotStable not observed — OBCluster may have finished before the first drift check (timing window missed)"
    fi

    # Single-restart: generation must advance by exactly 1
    local max_gen
    max_gen=$(grep -o "gen=[0-9]*" "$MONITOR_LOG" | sed 's/gen=//' | sort -n | tail -1)
    local expected_gen=$(( gen_before + 1 ))
    echo ""
    if [[ -n "$max_gen" && "$max_gen" -le "$expected_gen" ]]; then
        echo "PASS (single-restart): generation ${gen_before} → ${max_gen} — no spurious extra restarts"
    else
        echo "FAIL (single-restart): generation ${gen_before} → ${max_gen:-?} — expected at most ${expected_gen}"
    fi

    rm -f "$MONITOR_LOG"
    MONITOR_LOG=""
}

count_observer_crs() {
    # Only count observer CRs that are not in deleting state
    # Use grep -v to filter out observers with "deleting" status
    sudo kubectl get observer -n "$NAMESPACE" -l "${REF_CLUSTER_LABEL}=${OBCLUSTER_NAME}" --no-headers 2>/dev/null | grep -v "deleting" | wc -l | tr -d ' '
}

count_observer_crs_running() {
    sudo kubectl get observer -n "$NAMESPACE" -l "${REF_CLUSTER_LABEL}=${OBCLUSTER_NAME}" \
        -o jsonpath='{.items[*].status.status}' 2>/dev/null | grep -o running | wc -l | tr -d ' '
}

get_observer_count() {
    local ip
    ip=$(kubectl get pod -o wide -n "$NAMESPACE" | grep "${OBCLUSTER_NAME}-1-zone1" | awk '{print $6}' | head -1)
    if [[ -n "$ip" ]]; then
        mysql -uroot -h "$ip" -P 2881 -Doceanbase -p"$PASSWORD" -e 'select count(*) from __all_server where status="active";' -N 2>/dev/null
    else
        echo "0"
    fi
}

wait_for_no_deleting_observers() {
    local timeout="${1:-180}"
    local counter=0
    local max_counter=$((timeout / 5))
    echo "Waiting for all deleting observers to be cleaned up..."
    while true; do
        counter=$((counter + 1))
        local deleting_count
        # Count observers with "deleting" status
        deleting_count=$(sudo kubectl get observer -n "$NAMESPACE" -l "${REF_CLUSTER_LABEL}=${OBCLUSTER_NAME}" --no-headers 2>/dev/null | grep "deleting" | wc -l | tr -d ' ')
        
        if [[ "$deleting_count" -eq 0 ]]; then
            echo "No deleting observers found."
            return 0
        fi
        
        echo "  Still have $deleting_count observers in deleting state... ($counter/$max_counter)"
        
        if [[ $counter -ge $max_counter ]]; then
            echo "WARNING: Timeout waiting for deleting observers to clean up"
            return 1
        fi
        sleep 5
    done
}

get_current_observer_count() {
    count_observer_crs
}

reset_cluster_to_1_1_1() {
    local current_count
    current_count=$(get_current_observer_count)
    
    echo "Current observer count: $current_count"
    
    if [[ "$current_count" -eq 3 ]]; then
        echo "Cluster is already in 1-1-1 state (3 observers)."
        return 0
    fi
    
    echo "Cluster is in $current_count observers state, resetting to 1-1-1..."
    
    # Apply 1-1-1 configuration
    envsubst < "$TESTS_DIR/config/clusterManage/obcluster_template_1-1-1.yaml" | kubectl apply -f -
    
    # Wait for scale in to complete
    echo "Waiting for cluster to scale in to 1-1-1..."
    wait_observer_targets 3 3 300 || {
        echo "ERROR: Failed to scale in to 1-1-1"
        return 1
    }
    
    # Wait for deleting observers to be cleaned up
    wait_for_no_deleting_observers 180 || {
        echo "WARNING: Some observers still in deleting state after reset"
    }
    
    echo "Cluster successfully reset to 1-1-1 state."
    return 0
}

wait_observer_targets() {
    local want_total="$1"
    local want_db="$2"
    local timeout="${3:-180}"
    local counter=0
    echo "Waiting for OBServer CRs total=$want_total running=$want_total (DB active=$want_db when reachable)..."
    while true; do
        counter=$((counter + 1))
        local total running db
        total=$(count_observer_crs)
        running=$(count_observer_crs_running)
        db=$(get_observer_count)
        echo "  CR total=$total running=$running (want $want_total); DB active=$db (want $want_db)"
        if [[ "$total" == "$want_total" && "$running" == "$want_total" && "$db" == "$want_db" ]]; then
            echo "Observer targets satisfied."
            return 0
        fi
        if [[ $counter -ge $timeout ]]; then
            echo "Timeout waiting for observer targets"
            return 1
        fi
        sleep 5
    done
}

wait_obproxy_rollout() {
    local deploy="$1"
    local timeout="${2:-300s}"
    if [[ -z "$deploy" ]]; then
        echo "ERROR: OBProxy Deployment not found"
        return 1
    fi
    echo "Waiting for Deployment rollout: $deploy"
    kubectl rollout status "deployment/$deploy" -n "$NAMESPACE" --timeout="$timeout"
}

test_scale_out() {
    echo "=== Testing RS_LIST change on scale out ==="

    local deploy rs_before observers_before gen_before
    deploy=$(get_obproxy_deployment_name)
    rs_before=$(get_rs_list)
    observers_before=$(get_observer_count)
    gen_before=$(get_deployment_generation)

    echo "Before scale out:"
    echo "  RS_LIST (from Pod): $rs_before"
    echo "  Observer count (DB): $observers_before"
    echo "  Deployment: $deploy generation=$gen_before"
    echo "  Debug - OBProxy status: $(get_obproxy_status)"

    echo "Scaling cluster to 2-2-2..."
    start_condition_monitor
    envsubst < "$TESTS_DIR/config/clusterManage/obcluster_template_2-2-2.yaml" | kubectl apply -f -

    wait_observer_targets 6 6 180 || true

    # Debug: Show Observer CRs
    echo ""
    echo "Observer CRs after scale out:"
    kubectl get observer -n "$NAMESPACE" -l "${REF_CLUSTER_LABEL}=${OBCLUSTER_NAME}" -o custom-columns=NAME:.metadata.name,STATUS:.status.status 2>/dev/null || echo "  (unable to get observers)"

    echo "Waiting for RS_LIST change AND rolling update (generation must change)..."
    local counter=0 timeout=120 rs_after gen_after
    while true; do
        counter=$((counter + 1))
        deploy=$(get_obproxy_deployment_name)
        rs_after=$(get_rs_list)
        gen_after=$(get_deployment_generation)

        echo "Checking... ($counter/$timeout)"
        echo "  RS_LIST (Pod): $rs_after"
        echo "  generation: $gen_after (was $gen_before)"

        local rs_changed=0 rollout_signal=0
        [[ -n "$rs_after" && "$rs_before" != "$rs_after" ]] && rs_changed=1
        [[ -n "$gen_after" && "$gen_after" != "$gen_before" ]] && rollout_signal=1

        if [[ "$rs_changed" -eq 1 && "$rollout_signal" -eq 1 ]]; then
            echo ""
            echo "SUCCESS: RS_LIST changed and Deployment spec changed (rolling update triggered)."
            echo "  RS before: $rs_before"
            echo "  RS after:  $rs_after"
            break
        fi

        if [[ $counter -ge $timeout ]]; then
            echo "WARNING: RS_LIST or rollout signal did not match within timeout (need RS change AND generation change)."
            echo ""
            echo "=== Debug Info on Timeout ==="
            echo "Deployment env vars:"
            kubectl get deployment -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" -o jsonpath='{.items[0].spec.template.spec.containers[0].env}' 2>/dev/null || echo "  Deployment not found"
            echo ""
            echo "Recent controller logs:"
            get_obproxy_controller_logs "3m"
            break
        fi
        sleep 5
    done

    if [[ -n "$deploy" ]]; then
        wait_obproxy_rollout "$deploy" 300s || echo "WARNING: rollout status did not succeed in time"
    fi

    stop_condition_monitor

    rs_after=$(get_rs_list)
    if verify_all_obproxy_pods_rs_list "$rs_after"; then
        echo "All Running OBProxy pods share the same RS_LIST."
    else
        echo "WARNING: not all OBProxy pods agree on RS_LIST after rollout."
    fi

    analyze_debounce "$gen_before" "scale-out"
}

test_scale_in() {
    echo ""
    echo "=== Testing RS_LIST change on scale in ==="

    local deploy rs_before gen_before
    deploy=$(get_obproxy_deployment_name)
    rs_before=$(get_rs_list)
    gen_before=$(get_deployment_generation)

    echo "Before scale in:"
    echo "  RS_LIST (Pod): $rs_before"
    echo "  Deployment: $deploy generation=$gen_before"

    echo "Scaling cluster back to 1-1-1..."
    start_condition_monitor
    envsubst < "$TESTS_DIR/config/clusterManage/obcluster_template_1-1-1.yaml" | kubectl apply -f -

    wait_observer_targets 3 3 180 || true

    # Debug: Show Observer CRs
    echo ""
    echo "Observer CRs after scale in:"
    kubectl get observer -n "$NAMESPACE" -l "${REF_CLUSTER_LABEL}=${OBCLUSTER_NAME}" -o custom-columns=NAME:.metadata.name,STATUS:.status.status 2>/dev/null || echo "  (unable to get observers)"

    echo "Waiting for RS_LIST change AND rolling update..."
    local counter=0 timeout=120 rs_after gen_after
    while true; do
        counter=$((counter + 1))
        deploy=$(get_obproxy_deployment_name)
        rs_after=$(get_rs_list)
        gen_after=$(get_deployment_generation)

        echo "Checking... ($counter/$timeout)"
        echo "  RS_LIST (Pod): $rs_after"
        echo "  generation: $gen_after (was $gen_before)"

        local rs_changed=0 rollout_signal=0
        [[ -n "$rs_after" && "$rs_before" != "$rs_after" ]] && rs_changed=1
        [[ -n "$gen_after" && "$gen_after" != "$gen_before" ]] && rollout_signal=1

        if [[ "$rs_changed" -eq 1 && "$rollout_signal" -eq 1 ]]; then
            echo ""
            echo "SUCCESS: RS_LIST changed and Deployment spec changed (rolling update triggered)."
            echo "  RS before: $rs_before"
            echo "  RS after:  $rs_after"
            break
        fi

        if [[ $counter -ge $timeout ]]; then
            echo "WARNING: RS_LIST or rollout signal did not match within timeout."
            echo ""
            echo "=== Debug Info on Timeout ==="
            echo "Deployment env vars:"
            kubectl get deployment -n "$NAMESPACE" -l "${OBPROXY_LABEL}=${OBPROXY_NAME}" -o jsonpath='{.items[0].spec.template.spec.containers[0].env}' 2>/dev/null || echo "  Deployment not found"
            echo ""
            echo "Recent controller logs:"
            get_obproxy_controller_logs "3m"
            break
        fi
        sleep 5
    done

    if [[ -n "$deploy" ]]; then
        wait_obproxy_rollout "$deploy" 300s || echo "WARNING: rollout status did not succeed in time"
    fi

    stop_condition_monitor

    rs_after=$(get_rs_list)
    if verify_all_obproxy_pods_rs_list "$rs_after"; then
        echo "All Running OBProxy pods agree on RS_LIST."
    else
        echo "WARNING: not all OBProxy pods agree on RS_LIST after rollout."
    fi

    analyze_debounce "$gen_before" "scale-in"
}

echo "=== OBProxy RS Change Test ==="
echo "NAMESPACE: $NAMESPACE"
echo "OBCLUSTER_NAME: $OBCLUSTER_NAME"
echo "OBPROXY_NAME: $OBPROXY_NAME"

# Pre-check 1: Ensure no observers in deleting state
echo ""
echo "=== Pre-check 1: Ensuring no observers in deleting state ==="
wait_for_no_deleting_observers 60 || echo "WARNING: Some observers still in deleting state, continuing anyway..."

# Pre-check 2: Check current cluster state and reset to 1-1-1 if needed
echo ""
echo "=== Pre-check 2: Checking initial cluster state ==="
current_observer_count=$(get_current_observer_count)
echo "Current observer count: $current_observer_count"

if [[ "$current_observer_count" -eq 3 ]]; then
    echo "Cluster is already in 1-1-1 state, ready to start test."
elif [[ "$current_observer_count" -eq 6 ]]; then
    echo "Cluster is in 2-2-2 state, resetting to 1-1-1..."
    if ! reset_cluster_to_1_1_1; then
        echo "ERROR: Failed to reset cluster to 1-1-1 state. Aborting test."
        exit 1
    fi
else
    echo "WARNING: Cluster is in unexpected state with $current_observer_count observers."
    echo "Attempting to reset to 1-1-1..."
    if ! reset_cluster_to_1_1_1; then
        echo "ERROR: Failed to reset cluster to 1-1-1 state. Aborting test."
        exit 1
    fi
fi

# Final check: Verify cluster is in 1-1-1 state
echo ""
echo "=== Final verification: Ensuring cluster is in 1-1-1 state ==="
final_count=$(get_current_observer_count)
if [[ "$final_count" -ne 3 ]]; then
    echo "ERROR: Cluster is not in 1-1-1 state (has $final_count observers). Aborting test."
    exit 1
fi
echo "Cluster is ready in 1-1-1 state with 3 observers."

test_scale_out
test_scale_in

echo ""
echo "=== RS Change Test Completed ==="