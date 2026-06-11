#!/bin/bash

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TESTS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# load all the parameters in setup.sh
source "$TESTS_DIR/setup.sh"
source "$TESTS_DIR/util.sh"
source "$TESTS_DIR/env.sh"
source "$TESTS_DIR/case_p0_OBProxy/env_vars.sh"

# OBProxy label selector
OBPROXY_LABEL="obproxy.oceanbase.com/obproxy"

get_deployment_generation() {
    kubectl get deployment -n $NAMESPACE -l ${OBPROXY_LABEL}=${OBPROXY_NAME} -o jsonpath='{.items[0].metadata.generation}' 2>/dev/null
}

get_pod_restarts() {
    kubectl get pods -n $NAMESPACE -l ${OBPROXY_LABEL}=${OBPROXY_NAME} -o jsonpath='{.items[*].status.containerStatuses[0].restartCount}' 2>/dev/null | tr ' ' '\n' | awk '{sum+=$1} END {print sum}'
}

get_ready_replicas() {
    kubectl get obproxy $OBPROXY_NAME -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null
}

get_desired_replicas() {
    kubectl get obproxy $OBPROXY_NAME -n $NAMESPACE -o jsonpath='{.status.replicas}' 2>/dev/null
}

# Test rolling update when image changes
test_image_update() {
    echo "=== Testing OBProxy rolling update on image change ==="
    
    local gen_before=$(get_deployment_generation)
    local image_before=$(kubectl get obproxy $OBPROXY_NAME -n $NAMESPACE -o jsonpath='{.spec.image}')
    local pods_before=$(kubectl get pods -n $NAMESPACE -l ${OBPROXY_LABEL}=${OBPROXY_NAME} -o name | wc -l)
    
    echo "Before image update:"
    echo "  Image: $image_before"
    echo "  Generation: $gen_before"
    echo "  Pod count: $pods_before"
    
    # Update to a different image
    local new_image="oceanbase/obproxy-ce:4.3.2.0-1"
    echo "Updating image to: $new_image"
    kubectl patch obproxy $OBPROXY_NAME -n $NAMESPACE --type=merge -p "{\"spec\":{\"image\":\"$new_image\"}}"
    
    # Wait for rolling update
    local counter=0
    local timeout=180
    while true; do
        counter=$((counter+1))
        local gen_after=$(get_deployment_generation)
        local ready=$(get_ready_replicas)
        local desired=$(get_desired_replicas)
        local current_image=$(kubectl get obproxy $OBPROXY_NAME -n $NAMESPACE -o jsonpath='{.status.image}')
        
        echo "Waiting for rolling update... ($counter/$timeout)"
        echo "  Generation: $gen_after (was $gen_before)"
        echo "  Ready: $ready/$desired"
        echo "  Current image: $current_image"
        
        if [[ "$gen_after" != "$gen_before" && "$ready" == "$desired" && -n "$ready" ]]; then
            echo ""
            echo "SUCCESS: Rolling update completed!"
            echo "  Generation changed: $gen_before -> $gen_after"
            echo "  All pods ready: $ready/$desired"
            break
        fi
        
        if [ $counter -eq $timeout ]; then
            echo "Rolling update did not complete within timeout"
            kubectl describe obproxy $OBPROXY_NAME -n $NAMESPACE
            kubectl get pods -n $NAMESPACE -l ${OBPROXY_LABEL}=${OBPROXY_NAME}
            break
        fi
        sleep 5s
    done
}

# Test replica scaling
test_replica_scale() {
    echo ""
    echo "=== Testing OBProxy replica scaling ==="
    
    local replicas_before=$(kubectl get obproxy $OBPROXY_NAME -n $NAMESPACE -o jsonpath='{.spec.replicas}')
    echo "Replicas before: $replicas_before"
    
    # Scale to 3
    local new_replicas=3
    echo "Scaling to $new_replicas replicas..."
    kubectl patch obproxy $OBPROXY_NAME -n $NAMESPACE --type=merge -p "{\"spec\":{\"replicas\":$new_replicas}}"
    
    local counter=0
    local timeout=120
    while true; do
        counter=$((counter+1))
        local ready=$(get_ready_replicas)
        local desired=$(get_desired_replicas)
        
        echo "Waiting for scale... ready: $ready/$desired"
        
        if [[ "$ready" == "$new_replicas" && "$desired" == "$new_replicas" ]]; then
            echo ""
            echo "SUCCESS: Scale to $new_replicas completed!"
            break
        fi
        
        if [ $counter -eq $timeout ]; then
            echo "Scale did not complete within timeout"
            break
        fi
        sleep 3s
    done
    
    # Scale back to original
    echo ""
    echo "Scaling back to $replicas_before replicas..."
    kubectl patch obproxy $OBPROXY_NAME -n $NAMESPACE --type=merge -p "{\"spec\":{\"replicas\":$replicas_before}}"
    
    counter=0
    while true; do
        counter=$((counter+1))
        local ready=$(get_ready_replicas)
        local desired=$(get_desired_replicas)
        
        echo "Waiting for scale back... ready: $ready/$desired"
        
        if [[ "$ready" == "$replicas_before" && "$desired" == "$replicas_before" ]]; then
            echo ""
            echo "SUCCESS: Scale back to $replicas_before completed!"
            break
        fi
        
        if [ $counter -eq $timeout ]; then
            echo "Scale back did not complete within timeout"
            break
        fi
        sleep 3s
    done
}

# Test pod restart (delete pod and verify it comes back)
test_pod_restart() {
    echo ""
    echo "=== Testing OBProxy pod restart ==="
    
    local pods=$(kubectl get pods -n $NAMESPACE -l ${OBPROXY_LABEL}=${OBPROXY_NAME} -o name)
    local pod_to_delete=$(echo "$pods" | head -1)
    
    echo "Deleting pod: $pod_to_delete"
    kubectl delete $pod_to_delete -n $NAMESPACE
    
    local counter=0
    local timeout=60
    while true; do
        counter=$((counter+1))
        local ready=$(get_ready_replicas)
        local desired=$(get_desired_replicas)
        
        echo "Waiting for pod recovery... ready: $ready/$desired"
        
        if [[ "$ready" == "$desired" && -n "$ready" ]]; then
            echo ""
            echo "SUCCESS: Pod recovered successfully!"
            break
        fi
        
        if [ $counter -eq $timeout ]; then
            echo "Pod did not recover within timeout"
            break
        fi
        sleep 3s
    done
}

echo "=== OBProxy Restart/Rolling Update Test ==="
echo "NAMESPACE: $NAMESPACE"
echo "OBPROXY_NAME: $OBPROXY_NAME"

test_image_update
test_replica_scale
test_pod_restart

echo ""
echo "=== Restart/Rolling Update Test Completed ==="