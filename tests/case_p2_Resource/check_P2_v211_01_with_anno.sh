#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
#source case_p2_Resource/env_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export OB_CLUSTER_ID=$(( (RANDOM % 10000) + 1000 ))
    export OB_CLUSTER_ID_NEW=$(( (RANDOM % 10000) + 1000 ))
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=test$SUFFIX
    export OBCLUSTER_NAME_NEW=testnew$SUFFIX
    export OB_ROOT_SECRET=sc-root-$SUFFIX
    export OBTENANT_NAME=tenant$SUFFIX
    kubectl create namespace $NAMESPACE
    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
    echo $PASSWORD
}
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}
 
run() {
    local output_yaml=$1
    envsubst < ./config/2.1.1_test/$output_yaml | kubectl apply -f -
}

delete() {
    local output_yaml=$1
    envsubst < ./config/2.1.1_test/$output_yaml | kubectl delete -f -
}

check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1 |awk -F' ' '{print $2}'| awk 'NR==1'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "1/1"  && -n "$ip" && $crd_obcluster = "status:running" ]];then
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

check_resource_new_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_NEW-$OB_CLUSTER_ID_NEW-zone1 |awk -F' ' '{print $2}'| awk 'NR==1'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_NEW-$OB_CLUSTER_ID_NEW-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME_NEW -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "1/1"  && -n "$ip" && $crd_obcluster = "status:running" ]];then
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
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1|awk -F' ' '{print $11}'| awk 'NR==1'`
        echo $server_1_zone1
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

check_in_new_obcluster() {
    counter=0
    timeout=100  
    OBSERVER_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_NEW-$OB_CLUSTER_ID_NEW-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1|awk -F' ' '{print $11}'| awk 'NR==1'`
        echo $server_1_zone1
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

check_pvc() {
    counter=0
    timeout=100  
    PVC_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_NEW-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob pvc'
        counter=$((counter+1))
        pvc_data_file=$(kubectl get pvc -n "$NAMESPACE" | grep "$OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1-.*-data-file")
	pvc_data_log=$(kubectl get pvc -n "$NAMESPACE" | grep "$OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1-.*-data-log")
	pvc_log=$(kubectl get pvc -n "$NAMESPACE" | awk "/$OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1-.*-log/ && !/$OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1-.*-data-log/")
	pvc_data_file_no_anno=$(kubectl get pvc -n "$NAMESPACE" | grep "$OBCLUSTER_NAME_NEW-$OB_CLUSTER_ID_NEW-zone1-.*-data-file")
	pvc_data_log_no_anno=$(kubectl get pvc -n "$NAMESPACE" | grep "$OBCLUSTER_NAME_NEW-$OB_CLUSTER_ID_NEW-zone1-.*-data-log")
	pvc_log_no_anno=$(kubectl get pvc -n "$NAMESPACE" | grep "$OBCLUSTER_NAME_NEW-$OB_CLUSTER_ID_NEW-zone1-.*[^-data]-log")
	echo "pvc_data_file $pvc_data_file pvc_data_log $pvc_data_log pvc_log $pvc_log pvc_data_file_no_anno $pvc_data_file_no_anno pvc_data_log_no_anno $pvc_data_log_no_anno pvc_log_no_anno $pvc_log_no_anno"
        if [[ -n "$pvc_data_file" && -n "$pvc_data_log" && -n "$pvc_log" && -z "$pvc_data_file_no_anno" && -z "$pvc_data_log_no_anno" && -z "$pvc_log_no_anno" ]]
        then
            PVC_ACTIVE='true'
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

export_to_file() {
    local output_file="case_p2_Resource/env_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBTENANT_NAME="$OBTENANT_NAME"
export OB_CLUSTER_ID="$OB_CLUSTER_ID"
export OB_CLUSTER_ID_NEW="$OB_CLUSTER_ID_NEW"
export OBCLUSTER_NAME_NEW="$OBCLUSTER_NAME_NEW"
EOF
    echo "Environment variables have been exported to $output_file"
}

validate() {
    run "case1_cluster_with_anno.yaml"
    run "case1_cluster_without_anno.yaml"
    echo 'do validate'
    check_resource_running
    check_resource_new_running
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	check_in_new_obcluster
        delete "case1_cluster_with_anno.yaml"
	delete "case1_cluster_without_anno.yaml"
	check_pvc
        if [[ $PVC_ACTIVE == 'false' ]]; then
            echo "case failed"
            cleanup
        else
            echo "case passed"
        fi
    fi
}

prepare
export_to_file
validate

