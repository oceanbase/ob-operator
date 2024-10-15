#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p4_ResourceModification/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    export RESCUE_RESET=rescue-reset-$SUFFIX
    export RESCUE_RESET_NOOP=rescue-reset-noop-$SUFFIX
    echo $PASSWORD
} 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}

run() {
    local template_file=$1
    local action=${2:-apply}  
    echo $OBCLUSTER_NAME
    case $action in
        apply)
            envsubst < ./config/2.2.0_test/$template_file | kubectl apply -f - ;;
        create)
            envsubst < ./config/2.2.0_test/$template_file | kubectl create -f - ;;
        *)
            return 1 ;;
    esac
}
 
check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone1" | awk -v line="1" 'NR==line{print \$2}'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "1/1" && -n "$ip" && $crd_obcluster = "status:running" ]];then
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
    counter=0
    timeout=100  
    OBSERVER_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1| awk -v line="1" 'NR==line{print \$11}'`
        if [[ $server_1_zone1 == "ACTIVE" ]]
        then
            echo "server_1_zone1 $server_1_zone1"
            OBSERVER_ACTIVE='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
	    echo "case failed"
            break
        fi
        sleep 3s
    done
}
 
check_obtenant_not_running() {
    counter=0
    timeout=100  
    OBTENANT_RUNNING='false'
    while true; do
        echo 'check obtenant resource'
	counter=$((counter+1))
	result=$(kubectl get obtenant $OBTENANT_NAME -n $NAMESPACE | grep "$OBTENANT_NAME" | awk -F' ' '{print $2}')
	output=$(kubectl describe obtenant $OBTENANT_NAME -n $NAMESPACE)
	echo "tenant $OBTENANT_NAME status $result"
	filtered_output=$(echo "$output" | grep "pool memory size is less than min_full_resource_pool_memory")
	echo $filtered_output
	if echo "$output" | grep -q "pool memory size is less than min_full_resource_pool_memory" && [ "$result" = "failed" ]; then
            OBTENANT_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "obtenant resource still not running"
            break
        fi
        sleep 3s
    done
}

check_rescue_running() { 
    local rescue_status=$1
    counter=0
    timeout=100  
    RESCUE_RUNNING='false'
    while true; do
        echo 'check obtenant rescue  resource'
        counter=$((counter+1))
	result=$(kubectl get obtenant $OBTENANT_NAME -n $NAMESPACE | grep "$OBTENANT_NAME" | awk -F' ' '{print $2}')
	echo "$result"
	if [ "$result" = "$rescue_status" ]; then
            RESCUE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "obtenant rescue  resource still not running"
            break
        fi
        sleep 3s
    done
}

validate() {
    run "obcluster_template_1-1-1.yaml"
    echo 'do validate'
    check_resource_running 
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster 
	run "obtenant_basic_1_1_1.yaml"
	check_obtenant_not_running
	run "rescue_reset.yaml" create
	check_rescue_running "creating" 
	run "rescue_reset_noop.yaml" create
	check_rescue_running "noop"
        if [[ $RESCUE_RUNNING == 'false' ]]; then
            echo "case failed"
	    #cleanup
	    #prepare
        else
            echo "case passed"
        fi
    fi
}

prepare
validate 

