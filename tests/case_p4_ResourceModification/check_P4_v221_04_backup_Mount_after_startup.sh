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

check_modifyback_running() {
    counter=0
    timeout=100  
    MODIYFBACK_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
	yaml_output=$(kubectl get pods -n $NS_DEFAULT -o yaml)
	if echo "$yaml_output" | grep -qE "path: $NFS_BASE_PATH" && echo "$yaml_output" | grep -qE "server: $NFS_SERVER" ; then
            MODIYFBACK_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            break
        fi
        sleep 3s
    done
}
 
validate() {
    run "obcluster_no_backup.yaml"
    check_resource_running  $OBCLUSTER_RECOVERY $NS_DEFAULT $CLUSTER_ID_RECOVERY
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
	echo "case failed"
    else
        run "obcluster_with_backup.yaml"
	check_resource_running  $OBCLUSTER_RECOVERY $NS_DEFAULT $CLUSTER_ID_RECOVERY
	check_modifyback_running
        if [[ $MODIYFBACK_RUNNING == 'false' ]]; then
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
