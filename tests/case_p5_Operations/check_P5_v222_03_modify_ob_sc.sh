#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p5_Operations/env_vars.sh
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}

prepare() {
    export LOCAL_PATH_NEW=local-path-new$SUFFIX 
#    kubectl create namespace $NAMESPACE
#    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
#    kubectl create serviceaccount $NAMESPACE -n $NAMESPACE
}

export_to_file() {
    local output_file="case_p5_Operations/env_vars.sh"
    {
        echo "export LOCAL_PATH_NEW=\"$LOCAL_PATH_NEW\""
    } >> "$output_file"
    echo "Environment variables have been appended to $output_file"
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.2.2_test/$template_file | kubectl apply -f -
}

check_resource_running() {
    counter=0
    timeout=200  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $2}'| awk 'NR==1'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
	crd_observer=`kubectl get observer -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
	crd_zone=`kubectl get obzone -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "1/1" && -n "$ip" && $crd_obcluster = "status:running" && $crd_obcluster = "status:running" && $crd_zone = "status:running"  ]];then
            echo "pod_1_zone1 is $pod_1_zone1 ready"
            echo "svc is $ip ready"
            echo "crd_obcluster $crd_obcluster crd_obcluster $crd_obcluster crd_zone $crd_zone"
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
        echo $server_1_zone1
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

validate() {
    run "new_sc.yaml"
    run "obcluster_template_1.yaml"
    echo 'do validate'
    check_resource_running
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	run "obcluster_template_new_sc.yaml"
	sleep 3s
	check_resource_running
        if [[ $RESOURCE_RUNNING == 'false' ]]; then
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
export_to_file
