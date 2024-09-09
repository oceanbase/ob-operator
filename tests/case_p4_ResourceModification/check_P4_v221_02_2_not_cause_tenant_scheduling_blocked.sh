#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p4_v220_and_v211/env_221_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    echo $PASSWORD
    echo $SUFFIX
    export OBCLUSTER_RECOVERY=obcluster-recovery-$SUFFIX
    export NS_DEFAULT=$NAMESPACE_PREFIX-default-$SUFFIX
    export CLUSTER_ID_RECOVERY=$(( (RANDOM % 9000) + 1000 ))
    export OB_ROOT_SECRET=ob-root-sc-$SUFFIX
    export SA_DEFAULT=sa-default-$SUFFIX
    export OBCLUSTER_DEFAULT_ONE=obcluster-default-1-$SUFFIX
    export NS_DEFAULT=$NAMESPACE_PREFIX-default-$SUFFIX
    export CLUSTER_ID_DEFAULT_ONE=$(( (RANDOM % 9000) + 1000 ))
    export TENANT_WITH_THREE_NODE=tenantthreenode$SUFFIX
    export TENANT_DEFAULT_ONE=tenantdefault1$SUFFIX
    kubectl create namespace $NS_DEFAULT
    kubectl create sa $SA_DEFAULT -n $NS_DEFAULT
    create_pass_secret $NS_DEFAULT $OB_ROOT_SECRET $PASSWORD
    
}

export_to_file() {
    local output_file="case_p4_v220_and_v211/env_221_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export OBCLUSTER_RECOVERY="$OBCLUSTER_RECOVERY"
export NS_DEFAULT="$NS_DEFAULT"
export CLUSTER_ID_RECOVERY="$CLUSTER_ID_RECOVERY"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export SA_DEFAULT="$SA_DEFAULT"
export OBCLUSTER_DEFAULT_TWO="$OBCLUSTER_DEFAULT_TWO"
export CLUSTER_ID_DEFAULT_TWO="$CLUSTER_ID_DEFAULT_TWO"
export TENANT_WITH_THREE_NODE="$TENANT_WITH_THREE_NODE"
EOF
    echo "Environment variables have been exported to $output_file"
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NS_DEFAULT 
}

run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.2.1_test/$template_file | kubectl apply -f -
#    envsubst < ./config/2.2.1_test/$template_file 
}

check_resource_running() {
    local obcluster_name=$1
    local namespace=$2
    local cluster_id=$3
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod  -o wide -n $namespace | grep $obcluster_name-$cluster_id-zone1 |awk -F' ' '{print $2}'| awk 'NR==1'`
        ip=`kubectl get pod  -o wide -n $namespace | grep $obcluster_name-$cluster_id-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $obcluster_name -n $namespace  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "1/1" && -n "$ip" && $crd_obcluster = "status:running" ]];then
            echo "pod_1_zone1 is $pod_1_zone1 ready"
            echo "svc is $ip ready"
            echo "crd_obcluster $crd_obcluster"
	    echo $pod_1_zone1
            RESOURCE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
 	    exit 1
            break
        fi
        sleep 3s
    done
}

check_obtenant_running() {
    local obtenant_name=$1
    local namespace=$2

    counter=0
    timeout=100  
    OBTENANT_RUNNING='false'
    while true; do
        echo 'check obtenant resource'
        counter=$((counter+1))
        crd_obtenant=`kubectl get obtenant $obtenant_name -n $namespace| grep "$obtenant_name"|awk -F' ' '{print $2}'`
        if [[ $crd_obtenant = "running" ]];then
            echo "crd_obtenant is $crd_obtenant ready"
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
 
kill_two_pod() {
    local obcluster_name=$1
    local obcluster_id=$2
    local namespace=$3
    pod1=`kubectl get pods -n $namespace -l ref-obcluster=$obcluster_name | grep  $obcluster_name-$obcluster_id-zone2 | awk '{print $1}'`
    pod2=`kubectl get pods -n $NS_DEFAULT -l ref-obcluster=$OBCLUSTER_RECOVERY | grep  $OBCLUSTER_RECOVERY-$CLUSTER_ID_RECOVERY-zone3 | awk '{print $1}'`
    kubectl delete pods $pod1 $pod2 -n $NS_DEFAULT
}

check_both_tenant_active() {
    local obtenant1_name=$1
    local ns_one=$2
    local obtenant2_name=$3
    local ns_two=$4
    counter=0
    timeout=100  
    OBTENANT_BOTH_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        obtenant1=`kubectl get obtenant $obtenant1_name -n $ns_one | grep -w $obtenant1_name | awk '{print $2}'`
        obtenant2=`kubectl get obtenant $obtenant2_name -n $ns_two | grep -w $obtenant2_name | awk '{print $2}'`
	if [[ $obtenant1 = "running" && $obtenant2 = "running"   ]]; then
            OBTENANT_BOTH_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            break
        fi
        sleep 3s
    done

}

three_operate() {
    run "tenant_service_3_nodes_2.yaml"  
    kill_two_pod $OBCLUSTER_RECOVERY $CLUSTER_ID_RECOVERY $NS_DEFAULT
    run "tenant_ns_default_upgrade.yaml"
}

validate() {
    run "obcluster_service_3_nodes.yaml"
    run "obcluster_ns_default.yaml"
    check_resource_running  $OBCLUSTER_RECOVERY $NS_DEFAULT $CLUSTER_ID_RECOVERY
    check_resource_running $OBCLUSTER_DEFAULT_ONE $NS_DEFAULT $CLUSTER_ID_DEFAULT_ONE
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
	echo "case failed"
    else
        run "tenant_service_3_nodes.yaml"
	run "tenant_ns_default.yaml"
	check_obtenant_running $TENANT_WITH_THREE_NODE $NS_DEFAULT
        check_obtenant_running $TENANT_DEFAULT_ONE $NS_DEFAULT
        three_operate
        check_both_tenant_active $TENANT_DEFAULT_ONE $NS_DEFAULT $TENANT_WITH_THREE_NODE $NS_DEFAULT
        if [[ $OBTENANT_BOTH_RUNNING == 'false' ]]; then
            echo "case failed"
        else
            echo "case passed" 
        fi
    fi 
}

date
prepare
export_to_file
validate
cleanup
date
