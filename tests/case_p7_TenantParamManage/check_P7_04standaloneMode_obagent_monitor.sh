#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
 
# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=test$SUFFIX
    export OB_ROOT_SECRET=sc-root-$SUFFIX
    export OBTENANT_NAME=tenant$SUFFIX
    kubectl create namespace "$NAMESPACE" || return 1
    trap 'cleanup || exit 1' EXIT
    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
}
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.3.1_test/$template_file | kubectl apply -f -
}

check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $2}'| awk 'NR==1'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "2/2" && -n "$ip" && $crd_obcluster = "status:running" ]];then
            echo "pod_1_zone1 is $pod_1_zone1 ready"
            echo "svc is $ip ready"
            echo "crd_obcluster $crd_obcluster"
            RESOURCE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
            break
        fi
        sleep 3s
    done
}
 
check_in_obcluster() {
    local ip zones attempt
    OBSERVER_ACTIVE=false
    for ((attempt=0; attempt<100; attempt++)); do
        ip=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath='{.items[0].status.podIP}') || return 1
        zones=$(MYSQL_PWD="$PASSWORD" mysql --connect-timeout=3 --batch --skip-column-names -h "$ip" -P2881 -uroot -Doceanbase \
            -e "SELECT CONCAT(ZONE,'@',SVR_IP) FROM DBA_OB_SERVERS WHERE STATUS='ACTIVE' ORDER BY ZONE") || zones=''
        if [[ $zones == zone1@127.0.0.1 ]]; then
            OBSERVER_ACTIVE=true
            printf 'ACTIVE SQL zones: %s\n' "${zones//$'\n'/,}"
            return 0
        fi
        sleep 3
    done
    return 1
}
 
export_to_file() {
    local output_file="case_p7_TenantParamManage/env_vars.sh"
    cat <<EOF > "$output_file"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBTENANT_NAME="$OBTENANT_NAME"
EOF
    echo "Environment variables have been exported to $output_file"
}

validate() {
    check_resource_running
    if [[ $RESOURCE_RUNNING != true ]] || ! check_in_obcluster || ! check_obagent_metrics "$NAMESPACE" "$OBCLUSTER_NAME" standalone 1; then
        echo "case failed"
        return 1
    fi
    echo "case passed"
}

prepare || exit 1
export_to_file
run "cluster_standalone.yaml" || exit 1
validate
