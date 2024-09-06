#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p3_v212/env_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export OBCLUSTER_NAME_NEW=testnew$SUFFIX
}
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.1.2_test/$template_file | kubectl apply -f -
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
    counter=0
    timeout=100  
    OBSERVER_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1|awk -F' ' '{print $11}'| awk 'NR==1'`
	oceanbase_DBA_OB_SERVERS_ip=`mysql -uroot -h $ip -P 2881 -p$PASSWORD  -Doceanbase  -e 'SELECT * FROM oceanbase.DBA_OB_SERVERS;'|grep 127.0.0.1|awk -F' ' '{print $1}'| awk 'NR==1'` 
        echo $server_1_zone1
	echo $oceanbase_DBA_OB_SERVERS_ip
        if [[ $server_1_zone1 == "ACTIVE"  && $oceanbase_DBA_OB_SERVERS_ip = "127.0.0.1" ]]
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

check_ob_scaled_up() {
    counter=0
    timeout=100  
    OB_SCALED_OP_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob scaled up'
	counter=$((counter+1))
	cpu=`kubectl get pods -n $NAMESPACE  -l ref-obcluster=$OBCLUSTER_NAME -o jsonpath='{.items[0].spec.containers[0].resources.requests.cpu}'`
	memory=`kubectl get pods -n $NAMESPACE -l ref-obcluster=$OBCLUSTER_NAME -o jsonpath='{.items[0].spec.containers[0].resources.requests.memory}'`
        if [[ $cpu = 3 && $memory = '12Gi' ]]
        then
            echo "server_1_zone1 $server_1_zone1"
            OB_SCALED_OP_ACTIVE='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource check ob scaled up still not running"
            echo "case failed"
            break
        fi
        sleep 3s
    done
}
 
validate() {
    run "cluster_standalone.yaml"
    echo 'do validate'
    check_resource_running
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	run "cluster_standalone_scaled_up.yaml"
	check_resource_running
	check_ob_scaled_up
        if [[ $OB_SCALED_OP_ACTIVE == 'false' ]]; then
            echo "case failed"
            #cleanup
        else
            echo "case passed"
        fi
    fi
}

validate

