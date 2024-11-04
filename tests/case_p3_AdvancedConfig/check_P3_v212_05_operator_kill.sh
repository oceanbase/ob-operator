#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p3_AdvancedConfig/env_vars.sh
# prepare create related resources to create obcluster
prepare() {
    export OBCLUSTER_NAME_THREE_NODE=testthreenode$SUFFIX
}
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME_THREE_NODE
    envsubst < ./config/2.1.2_test/$template_file | kubectl apply -f -
}

check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_THREE_NODE-1-zone1 |awk -F' ' '{print $2}'| awk 'NR==1'`
        pod_1_zone2=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_THREE_NODE-1-zone2 |awk -F' ' '{print $2}'| awk 'NR==1'`
        pod_1_zone3=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_THREE_NODE-1-zone3 |awk -F' ' '{print $2}'| awk 'NR==1'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_THREE_NODE-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME_THREE_NODE -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "2/2" && $pod_1_zone2 = "2/2" && $pod_1_zone3 = "2/2" && -n "$ip" && $crd_obcluster = "status:running" ]];then
            echo "pod_1_zone1 is $pod_1_zone1 ready"
            echo "pod_1_zone2 is $pod_1_zone2 ready"
            echo "pod_1_zone3 is $pod_1_zone3 ready"
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
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_THREE_NODE-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1|awk -F' ' '{print $11}'| awk 'NR==1'`
        server_1_zone2=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone2|awk -F' ' '{print $11}'| awk 'NR==1'`
        server_1_zone3=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone3|awk -F' ' '{print $11}'| awk 'NR==1'`
 
        echo $server_1_zone1
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
            break
        fi
        sleep 3s
    done
}

killPod() {
    pod_1_zone1_name=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_THREE_NODE-1-zone1 |awk -F' ' '{print $1}'| awk 'NR==1'`
    kubectl delete pod $pod_1_zone1_name -n $NAMESPACE
}
killoperator() {
    kubectl delete pod -n oceanbase-system -l control-plane=controller-manager
}

check_killedpod_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
	recover=`kubectl get observer -n $NAMESPACE |grep $OBCLUSTER_NAME_THREE_NODE-1-zone1|awk -F' ' '{print $4}'| awk 'NR==1'`
        echo 'check resource'
        counter=$((counter+1))
        if [[ $recover = "recover" ]];then
	    killoperator
            break
        fi
        if [ $counter -eq $timeout ]; then
            break
        fi
        sleep 3s
    done
}


validate() {
    run "cluster_three_nodes.yaml"
    echo 'do validate'
    check_resource_running
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	check_killedpod_running
	check_resource_running
	check_in_obcluster	
        if [[ $OBSERVER_ACTIVE == 'false' ]]; then
            echo "case failed"
            #cleanup
        else
            echo "case passed"
        fi
    fi
}

prepare
validate

