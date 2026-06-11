#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TESTS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$TESTS_DIR/setup.sh"
source "$TESTS_DIR/util.sh"
source "$TESTS_DIR/env.sh"

prepare() {
    export PASSWORD=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE="oceanbase-${SUFFIX}"
    export OBCLUSTER_NAME="test${SUFFIX}"
    export OB_ROOT_SECRET="sc-root-${SUFFIX}"
    export OBPROXY_IMAGE=${OBPROXY_IMAGE:-oceanbase/obproxy-ce:4.3.3.0-5}
    export OBPROXY_NAME="obproxy-${SUFFIX}"
    export PROXY_SYS_SECRET="sc-proxyro-${SUFFIX}"
    export OBPROXY_REPLICAS=2

    kubectl create namespace "$NAMESPACE"
    create_pass_secret "$NAMESPACE" "$OB_ROOT_SECRET" "$PASSWORD"
    create_pass_secret "$NAMESPACE" "$PROXY_SYS_SECRET" "$PASSWORD"
    echo "Prepared environment: SUFFIX=$SUFFIX NAMESPACE=$NAMESPACE"
}

create_obcluster() {
    echo "Creating OBCluster: $OBCLUSTER_NAME in $NAMESPACE"
    envsubst < "$TESTS_DIR/config/clusterManage/obcluster_template_1-1-1.yaml" | kubectl apply -f -
}

wait_obcluster_running() {
    local counter=0
    local timeout=200
    echo "Waiting for OBCluster to reach running status..."
    while true; do
        counter=$((counter + 1))
        local status
        status=$(kubectl get obcluster "$OBCLUSTER_NAME" -n "$NAMESPACE" \
            -o jsonpath='{.status.status}' 2>/dev/null)
        local ready_pods
        ready_pods=$(kubectl get pod -n "$NAMESPACE" --no-headers 2>/dev/null \
            | awk '$2=="1/1" && $3=="Running"' | wc -l | tr -d ' ')
        echo "  ($counter/$timeout) status=$status ready_pods=$ready_pods"
        if [[ "$status" == "running" && "$ready_pods" -ge 3 ]]; then
            echo "OBCluster is running."
            return 0
        fi
        if [[ $counter -ge $timeout ]]; then
            echo "ERROR: OBCluster not running after timeout"
            kubectl describe obcluster "$OBCLUSTER_NAME" -n "$NAMESPACE" | tail -20
            return 1
        fi
        sleep 5
    done
}

create_obproxy() {
    echo "Creating OBProxy: $OBPROXY_NAME"
    envsubst < "$TESTS_DIR/config/obproxyManage/obproxy_template.yaml" | kubectl apply -f -
}

wait_obproxy_running() {
    local counter=0
    local timeout=100
    echo "Waiting for OBProxy to be ready..."
    while true; do
        counter=$((counter + 1))
        local status ready desired
        status=$(kubectl get obproxy "$OBPROXY_NAME" -n "$NAMESPACE" \
            -o jsonpath='{.status.status}' 2>/dev/null)
        ready=$(kubectl get obproxy "$OBPROXY_NAME" -n "$NAMESPACE" \
            -o jsonpath='{.status.readyReplicas}' 2>/dev/null)
        desired=$(kubectl get obproxy "$OBPROXY_NAME" -n "$NAMESPACE" \
            -o jsonpath='{.status.replicas}' 2>/dev/null)
        echo "  ($counter/$timeout) status=$status ready=$ready/$desired"
        if [[ "$status" == "running" && -n "$ready" && "$ready" == "$desired" ]]; then
            echo "OBProxy is running: $ready/$desired ready"
            return 0
        fi
        if [[ $counter -ge $timeout ]]; then
            echo "ERROR: OBProxy not running after timeout"
            kubectl describe obproxy "$OBPROXY_NAME" -n "$NAMESPACE" | tail -20
            return 1
        fi
        sleep 3
    done
}

check_rs_list() {
    echo "=== Checking RS_LIST configuration ==="
    local obproxy_label="obproxy.oceanbase.com/obproxy=$OBPROXY_NAME"
    local pod
    pod=$(kubectl get pod -n "$NAMESPACE" -l "$obproxy_label" -o name 2>/dev/null | head -1)
    if [[ -n "$pod" ]]; then
        echo "Checking RS_LIST in $pod:"
        kubectl exec -n "$NAMESPACE" "$pod" -- env 2>/dev/null \
            | grep -E "^RS_LIST=" || echo "  RS_LIST env not found"
    else
        echo "No OBProxy pod found"
    fi
}

check_connectivity() {
    echo "=== Checking OBProxy connectivity ==="
    local obproxy_label="obproxy.oceanbase.com/obproxy=$OBPROXY_NAME"
    local svc_ip
    svc_ip=$(kubectl get svc -n "$NAMESPACE" -l "$obproxy_label" \
        -o jsonpath='{.items[0].spec.clusterIP}' 2>/dev/null)
    if [[ -n "$svc_ip" ]]; then
        echo "OBProxy Service IP: $svc_ip"
        local proxy_sys_password
        proxy_sys_password=$(kubectl get secret "$PROXY_SYS_SECRET" -n "$NAMESPACE" \
            -o jsonpath='{.data.password}' | base64 -d)
        mysql -h "$svc_ip" -P 2883 -uroot@proxysys -p"$proxy_sys_password" \
            -e "SHOW PROXYCONFIG;" 2>&1 | head -5 \
            || echo "Connection test completed (mysql client may not be available)"
    else
        echo "OBProxy service not found"
    fi
}

export_to_file() {
    local output_file="$SCRIPT_DIR/env_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBPROXY_NAME="$OBPROXY_NAME"
export OBPROXY_IMAGE="$OBPROXY_IMAGE"
export PROXY_SYS_SECRET="$PROXY_SYS_SECRET"
export OBPROXY_REPLICAS="$OBPROXY_REPLICAS"
EOF
    echo "Environment exported to $output_file"
}

cleanup() {
    echo "Cleaning up namespace $NAMESPACE..."
    kubectl delete namespace "$NAMESPACE" --ignore-not-found=true
}

echo "=== OBProxy P8 Environment Setup ==="

prepare
export_to_file

create_obcluster
if ! wait_obcluster_running; then
    echo "FAILED: OBCluster did not reach running state"
    exit 1
fi

create_obproxy
if ! wait_obproxy_running; then
    echo "FAILED: OBProxy did not reach running state"
    exit 1
fi

check_rs_list
check_connectivity

echo ""
echo "=== Environment ready ==="
echo "NAMESPACE:      $NAMESPACE"
echo "OBCLUSTER_NAME: $OBCLUSTER_NAME"
echo "OBPROXY_NAME:   $OBPROXY_NAME"
echo "case passed"
