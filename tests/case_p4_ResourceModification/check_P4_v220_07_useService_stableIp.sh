#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p4_v220_and_v211/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    export OBCLUSTER_NAME=testservice$SUFFIX
    echo $PASSWORD
} 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    rm -rf case_p4_v220_and_v211/env_vars.sh
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
        if [[ $server_1_zone1 == "ACTIVE"  ]]
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

check_service_ip() {
    counter=0
    timeout=100  
    SERVICE_IP_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob'
        counter=$((counter+1))
	cluster_ip=$(kubectl get svc -n $NAMESPACE -l ref-obcluster=$OBCLUSTER_NAME | grep $OBCLUSTER_NAME-1-zone1 | awk -F' ' '{print $3}' | awk 'NR==1')
	echo "Service Cluster IP: $cluster_ip"
	db_result=$(obclient -uroot -h $ip -P2881 -p$PASSWORD -Doceanbase -e "select * from oceanbase.DBA_OB_SERVERS;" | grep zone1 | awk -F' ' '{print $1}' | awk 'NR==1')
	echo "Database query result: $db_result"
	if [ "$cluster_ip" = "$db_result" ]; then
            echo "Cluster IP and DB query result match, continue execution"
	    pod_name=$(kubectl get pod -n $NAMESPACE -l ref-obcluster=$OBCLUSTER_NAME | grep $OBCLUSTER_NAME-1-zone1 | awk -F' ' '{print $1}' | awk 'NR==1')
	    echo "Pod name: $pod_name"
	    kubectl delete pod -n $NAMESPACE -l ref-obcluster=$OBCLUSTER_NAME
	    check_resource_running
	    check_in_obcluster
	    new_pod_name=$(kubectl get pod -n $NAMESPACE -l ref-obcluster=$OBCLUSTER_NAME | grep $OBCLUSTER_NAME-1-zone1 | awk -F' ' '{print $1}' | awk 'NR==1')
	    echo "New Pod name: $new_pod_name"
 	    new_cluster_ip=$(kubectl get svc -n $NAMESPACE -l ref-obcluster=$OBCLUSTER_NAME | grep $OBCLUSTER_NAME-1-zone1 | awk -F' ' '{print $3}' | awk 'NR==1')
	    echo "New service Cluster IP: $new_cluster_ip"
	    db_result=$(obclient -uroot -h $ip -P2881 -p$PASSWORD -Doceanbase -e "select * from oceanbase.DBA_OB_SERVERS;" | grep zone1 | awk -F' ' '{print $1}' | awk 'NR==1')
	    if [ "$cluster_ip" = "$new_cluster_ip" ]  && [ "$pod_name" = "$new_pod_name" ]; then
		SERVICE_IP_ACTIVE='true'
	    fi
	    break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
	    break
        fi
        sleep 3s
    done
}

validate() {
    run "obcluster_service.yaml"
    echo 'do validate'
    check_resource_running 
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster 
	check_service_ip
        if [[ $SERVICE_IP_ACTIVE == 'false' ]]; then
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
cleanup_obcluster
cleanup
