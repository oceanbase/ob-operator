#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p2_v211/env_vars.sh
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    rm -rf case_p2_v211/env_vars.sh
}
 
run() {
    local output_yaml=$1
    envsubst < ./config/2.1.1_test/$output_yaml | kubectl apply -f -
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

check_disk_resources() {
    counter=0
    timeout=100  
    PARAMETERS_ACTIVE='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    echo $ip
    echo $PASSWORD
    while true; do
        echo 'check ob disk_resources'
        counter=$((counter+1))
	disk_resources=`sudo du -h $(kubectl get pv -n $NAMESPACE -o jsonpath="{.items[?(@.metadata.name=='$(kubectl get pv -n $NAMESPACE| grep "$OBCLUSTER_NAME-$OB_CLUSTER_ID-zone1-.*-data-file" | awk '{print $1}' | head -n 1)')].spec.hostPath.path}") | tail -n 1 | awk '{print $1}'`
	echo "disk_resources $disk_resources"
        if [[ $disk_resources = "11G" ]]
        then
            PARAMETERS_ACTIVE='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "check parameters still not running"
            break
        fi
        sleep 3s
    done

}

export_to_file() {
    local output_file="case_p2_v211/env_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBTENANT_NAME="$OBTENANT_NAME"
export OB_CLUSTER_ID="$OB_CLUSTER_ID"
EOF
    echo "Environment variables have been exported to $output_file"
}


validate() {
    run "case5_cluster_extra_opt.yaml"
    echo 'do validate'
    check_resource_running
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	check_disk_resources
        if [[ $PARAMETERS_ACTIVE == 'false' ]]; then
            echo "case failed"
            cleanup
        else
            echo "case passed"
        fi
    fi
}

validate
cleanup
