#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
#source case_p0_ClusterAndTenant/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=test$SUFFIX
    export OB_ROOT_SECRET=sc-root-$SUFFIX
    export OBTENANT_NAME=tenant$SUFFIX
    export TOP_DELETE_POOLS=top-delete-pools-$SUFFIX
    export TOP_MODIFY_POOLS=top-modify-pools-$SUFFIX
    export TOP_SET_CONNECT_WHITE_LIST=top-set-connect-white-list-$SUFFIX
    export TOP_SET_UNIT_NUMBER=top-set-unit-number-$SUFFIX
    kubectl create namespace "$NAMESPACE" || return 1
    trap finish EXIT
    create_pass_secret "$NAMESPACE" "$OB_ROOT_SECRET" "$PASSWORD"
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace "$NAMESPACE"
}

finish() {
    local rc=$?
    trap - EXIT
    cleanup || rc=1
    exit "$rc"
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.3.0_test/$template_file | kubectl apply -f -
}

create() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.3.0_test/$template_file | kubectl create  -f -
}

 
check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone1" | awk -v line="1" 'NR==line{print \$2}'`
        pod_1_zone2=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone2" | awk -v line="1" 'NR==line{print \$2}'`
        pod_1_zone3=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone3" | awk -v line="1" 'NR==line{print \$2}'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        if [[ $pod_1_zone1 = "1/1" && $pod_1_zone2 = "1/1" && $pod_1_zone3 = "1/1" && -n "$ip" ]];then
            echo "pod_1_zone1 is $pod_1_zone1 ready"
            echo "pod_1_zone2 is $pod_1_zone2 ready"
            echo "pod_1_zone3 is $pod_1_zone3 ready"
            echo "svc is $ip ready"
            RESOURCE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
            return 1
        fi
        sleep 3s
    done
}

check_in_obcluster() {
    counter=0
    timeout=100  
    OBSERVER_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1| awk -v line="1" 'NR==line{print \$11}'`
        server_1_zone2=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone2| awk -v line="1" 'NR==line{print \$11}'`
        server_1_zone3=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone3| awk -v line="1" 'NR==line{print \$11}'`
        if [[ $server_1_zone1 == "ACTIVE" && $server_1_zone2 == "ACTIVE" && $server_1_zone3 == "ACTIVE"  ]]
        then
            echo "server_1_zone1 $server_1_zone1"
            echo "server_1_zone2 $server_1_zone2"
            echo "server_1_zone3 $server_1_zone3"
            OBSERVER_ACTIVE='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
	    echo "case failed"
            return 1
        fi
        sleep 3s
    done
}
 
check_obtenant_running() {
    # The P0 unit-count scenario allows 12 minutes for a tenant on this topology.
    local deadline=$((SECONDS + 720)) state=''
    while ((SECONDS < deadline)); do
        state=$(kubectl -n "$NAMESPACE" get "obtenant/$OBTENANT_NAME" \
            --request-timeout=5s -o jsonpath='{.status.status}') || state=unavailable
        if [[ $state == running ]]; then
            echo "crd_obtenant is $state ready"
            echo "case passed"
            return 0
        fi
        [[ ${state,,} != failed ]] || break
        sleep 3
    done
    echo "case failed waiting for tenant $OBTENANT_NAME to reach running (last state $state)" >&2
    return 1
}

check_operation() {
    local template_file=$1
    counter=0
    timeout=100
    OPERATION_ACTIVE='false'
    while true; do
        echo 'check obtenantoperation'
        counter=$((counter+1))
        obtenantoperation=`kubectl get obtenantoperation -n $NAMESPACE | grep $template_file |awk -F' ' '{print $3}'| tail -n 1`
	echo "obclusteroperation $obtenantoperation"
        if [[ $obtenantoperation = "SUCCESSFUL"  ]]
        then
            echo "obclusteroperation $obtenantoperation"
            OPERATION_ACTIVE='true'
            echo "case pass"
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
            echo "case failed"
            return 1
        fi
        sleep 3s
    done
}

export_to_file() {
    local output_file="case_p6_SceneSettings/env_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBTENANT_NAME="$OBTENANT_NAME"
EOF
    echo "Environment variables have been exported to $output_file"
}


validate() {
    # The same six-Observer topology is used by the P0 unit-count scenario.
    if ! kubectl -n "$NAMESPACE" wait --for=jsonpath='{.status.status}'=running "obcluster/$OBCLUSTER_NAME" --timeout=15m; then
        echo "case failed waiting for cluster $OBCLUSTER_NAME to reach running" >&2
        return 1
    fi
    if ! check_resource_running; then
        echo "case failed waiting for ready Observer Pods" >&2
        return 1
    fi
    check_in_obcluster || return 1
    echo "case passed"
}

prepare || exit 1
run "obcluster_template_2-2-2.yaml" || exit 1
validate || exit 1
run "obtenant_basic_2_2_2.yaml" || exit 1
check_obtenant_running || exit 1
create "top-set-unit-number.yaml" || exit 1
check_obtenant_running || exit 1
check_operation "$TOP_SET_UNIT_NUMBER" || exit 1
