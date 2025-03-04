#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p2_Resource/env_211_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export PASSWORD_ENCRYPTION=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=test$SUFFIX
    export OBCLUSTER_NAME_TWO=testtwo$SUFFIX
    export OBTENANT_NAME=tenant$SUFFIX
    export OB_ROOT_SECRET=ob-root--sc-$SUFFIX
    export OBTENANT_ROOT_SECRET=obtenant-root-sc-$SUFFIX
    export OBTENANT_STANDBY_SECRET=obtenant-sy-sc-$SUFFIX
    export BACKUP_OSS_SERECT=backup-oss-sc-$SUFFIX
    export BACKUP_OSS_NAME=backup-oss-$SUFFIX
    export OBTENANT_STANDBY_OSS=restoresstandbyoss$SUFFIX
    export OBTENANT_ROOT_OSS_SECRET=restores-root-oss-sc-$SUFFIX
    export OBTENANT_STANDBY_OSS_SECRET=restores-sy-oss-sc-$SUFFIX
    export OSS_ARCHIVE_PATH=archive-path-$SUFFIX
    export OSS_BACPUP_PATH=backup-path-$SUFFIX
    export OSS_ACCESS=oss-access-$SUFFIX
    export BACK_ENCRYPTION_SERECT=back-ecp-sc-$SUFFIX
    export CLUSTER_ID_ONE=$(( (RANDOM % 9000) + 1000 ))
    export CLUSTER_ID_TWO=$(( (RANDOM % 9000) + 1000 )) 
    export OP_UPGRADE_TENANT=op-upgrade-tenant-$SUFFIX
    kubectl create namespace $NAMESPACE
    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
    create_pass_secret $NAMESPACE $OBTENANT_ROOT_SECRET $PASSWORD
    create_pass_secret $NAMESPACE $OBTENANT_STANDBY_SECRET $PASSWORD
    create_pass_secret $NAMESPACE $BACKUP_OSS_SERECT $PASSWORD
    create_pass_secret $NAMESPACE $OBTENANT_ROOT_OSS_SECRET $PASSWORD
    create_pass_secret $NAMESPACE $OBTENANT_STANDBY_OSS_SECRET $PASSWORD
    create_pass_secret $NAMESPACE $BACK_ENCRYPTION_SERECT $PASSWORD_ENCRYPTION
    create_pass_oss_secret $NAMESPACE $OSS_ACCESS $OSS_AK $OSS_SK
    echo $PASSWORD
}


# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    rm -rf case_p2_Resource/env_211_vars.sh
}

run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.1.1_test/$template_file | kubectl apply -f -
#    envsubst < ./config/2.1.1_test/$template_file 
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

check_backup_running() {
    counter=0
    timeout=100  
    BACKUP_RUNNING='false'
    while true; do
        echo 'check backup resource'
        counter=$((counter+1))
	crd_obtenantbackuppolicies_ARCHIVE=`kubectl get obtenantbackuppolicies.oceanbase.oceanbase.com -n $NAMESPACE| grep "$BACKUP_OSS_NAME"|awk -F' ' '{print $2}'`
	crd_FULL=`kubectl get obtenantbackups -n $NAMESPACE| grep "full"|awk -F' ' '{print $3}'| head -n 1`
        echo "backup resource is $crd_obtenantbackuppolicies_ARCHIVE obtenantbackups $crd_FULL"
        if [[ $crd_obtenantbackuppolicies_ARCHIVE = "RUNNING" && $crd_FULL = "SUCCESSFUL" ]];then
            BACKUP_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "backup  resource still not running"
            break
        fi
        sleep 3s
    done
}

check_restore_running() {
    counter=0
    timeout=120  
    RESTORE_RUNNING='false'
    while true; do
        echo 'check restore resource'
	counter=$((counter+1))
        crd_primary_restore=`kubectl get obtenant -n $NAMESPACE | grep "$OBTENANT_STANDBY_OSS"|awk -F' ' '{print $2}'`
        crd_primary_obtenant=`kubectl get obtenantrestore -n $NAMESPACE| grep "$OBTENANT_STANDBY_OSS"|awk -F' ' '{print $2}'`
        echo "crd_primary_restore $crd_primary_restore  crd_primary_obtenant $crd_primary_obtenant"
        if [[ $crd_primary_restore = "running" && $crd_primary_obtenant = "SUCCESSFUL" ]];then
            RESTORE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "backup db resource still not running"
            break
        fi
        sleep 3s
    done
}

check_operation() {
    counter=0
    timeout=100  
    OPERATION_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME_TWO-$CLUSTER_ID_TWO-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check operation resource'
	counter=$((counter+1))
	op_upgrade_tenant=`kubectl get obtenantoperation -n $NAMESPACE |grep $OP_UPGRADE_TENANT |awk -F' ' '{print $3}'| awk 'NR==1'`
	tenant421version=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'SELECT * FROM oceanbase.DBA_OB_TENANTS;'|grep $OBTENANT_STANDBY_OSS|awk -F' ' '{print $25}'| awk 'NR==1'`
	echo $op_upgrade_tenant $tenant421version
	if [[ $op_upgrade_tenant = "SUCCESSFUL" && $tenant421version = "4.2.1.1" ]];then
	    OPERATION_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            break
        fi
        sleep 3s
    done
}

export_to_file() {
    local output_file="case_p2_Resource/env_211_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OBTENANT_NAME="$OBTENANT_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBTENANT_ROOT_SECRET="$OBTENANT_ROOT_SECRET"
export OBTENANT_STANDBY_SECRET="$OBTENANT_STANDBY_SECRET"
export BACKUP_OSS_SERECT="$BACKUP_OSS_SERECT"
export BACKUP_OSS_NAME="$BACKUP_OSS_NAME"
export OBTENANT_STANDBY_OSS="$OBTENANT_STANDBY_OSS"
export OBTENANT_ROOT_OSS_SECRET="$OBTENANT_ROOT_OSS_SECRET"
export OBTENANT_STANDBY_OSS_SECRET="$OBTENANT_STANDBY_OSS_SECRET"
export OSS_ACCESS="$OSS_ACCESS"
export OSS_UNTIL_TIMESTAMP="$OSS_UNTIL_TIMESTAMP"
export OSS_REPLOY_LOG_UNTIL="$OSS_REPLOY_LOG_UNTIL"
export OSS_ARCHIVE_PATH="$OSS_ARCHIVE_PATH"
export OSS_BACPUP_PATH="$OSS_BACPUP_PATH"
export PASSWORD_ENCRYPTION="$PASSWORD_ENCRYPTION"
export BACK_ENCRYPTION_SERECT="$BACK_ENCRYPTION_SERECT"
export CLUSTER_ID_ONE="$CLUSTER_ID_ONE"
export CLUSTER_ID_TWO="$CLUSTER_ID_TWO"
export OBCLUSTER_NAME_TWO="$OBCLUSTER_NAME_TWO"
export OP_UPGRADE_TENANT="$OP_UPGRADE_TENANT"
EOF
    echo "Environment variables have been exported to $output_file"
}

validate() {
    run "case3_cluster_420.yaml"
    run "case3_cluster_421.yaml"
    check_resource_running $OBCLUSTER_NAME $NAMESPACE $CLUSTER_ID_ONE 
    check_resource_running $OBCLUSTER_NAME_TWO $NAMESPACE $CLUSTER_ID_TWO
    sleep 20s
    echo 'do validate'
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
	run "case3_tenant_420.yaml"
	check_obtenant_running $OBTENANT_NAME $NAMESPACE
	run "case3_policy_420.yaml"
        check_backup_running
	sleep 60s
	run "case3_restore_tenant_421.yaml"
	check_restore_running
	run "case3_op_upgrade_tenant.yaml"
	check_operation
        if [[ $OPERATION_RUNNING == 'false' ]]; then
            echo "case failed"
	    #cleanup
	    #prepare
        else
            echo "case passed"
        fi
    fi
}

prepare
export_to_file
validate
cleanup
date
