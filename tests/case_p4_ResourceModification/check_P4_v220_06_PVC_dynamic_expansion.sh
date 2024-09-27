#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p4_ResourceModification/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    export OBCLUSTER_NAME=testexpandable$SUFFIX
    export LOCAL_PATH_EXPANDABLE=local-path-expandable$SUFFIX
    echo $PASSWORD
} 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}

cleanup_obcluster() {
    kubectl delete obcluster $OBCLUSTER_NAME -n  $NAMESPACE
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.2.0_test/$template_file | kubectl apply -f -
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

check_pvc_in_obcluster() {
    counter=0
    timeout=100  
    OBSERVER_PVC_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob pvc'
        counter=$((counter+1))
	zone_status=$(kubectl get obzone -n $NAMESPACE -l ref-obcluster="$OBCLUSTER_NAME" | grep "$OBCLUSTER_NAME-1-zone1" | awk '{print $2}')
	observer_status=$(kubectl get observer -n $NAMESPACE -l ref-obcluster="$OBCLUSTER_NAME" | grep "$OBCLUSTER_NAME-1-zone1" | awk '{print $3}')
	obcluster_status=$(kubectl get obcluster -n $NAMESPACE | grep "$OBCLUSTER_NAME" | awk '{print $2}')
	pvc_storage=$(kubectl get pvc -n $NAMESPACE -l ref-obcluster="$OBCLUSTER_NAME" -o jsonpath='{.items[*].spec.resources.requests.storage}')
	echo "zone_status $zone_status observer_status $observer_status obcluster_status $obcluster_status pvc_storage $pvc_storage"
	if [ "$zone_status" = "running" ] && [ "$pvc_storage" = "61Gi 60Gi 20Gi" ] && [ "$observer_status" = "running" ] && [ "$obcluster_status" = "running" ]; then
            OBSERVER_PVC_ACTIVE='true'
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
 

validate() {
    run "sc_local_expandable.yaml"
    run "obcluster_expandable_1-1-1.yaml"
    echo 'do validate'
    check_resource_running 
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	run "obcluster_expanded_1-1-1.yaml"
	check_pvc_in_obcluster
        if [[ $OBSERVER_PVC_ACTIVE == 'false' ]]; then
            echo "case failed"
	    #cleanup
	    #prepare
        else
            echo "case passed"
        fi
    fi
}

prepare
cleanup_obcluster
validate 
cleanup_obcluster
