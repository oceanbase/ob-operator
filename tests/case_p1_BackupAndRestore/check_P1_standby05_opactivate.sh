#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p1_BackupAndRestore/env_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export OBTENANT_STANDBY=tenantstandy$SUFFIX
    export OP_ACTIVATE=op-acvivate$SUFFIX 
}

 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    ssh ${USER_NAME}@${NFS_SERVER} "sudo rm -rf ${NFS_BASE_PATH}/*"
}

cleanrestore() {
    kubectl delete obtenant $OBTENANT_STANDBY  -n $NAMESPACE
}

cleanoperation() {
    kubectl delete obtenantoperation $OP_ACTIVATE -n $NAMESPACE
}

run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/standby_tests/$template_file | kubectl apply -f -
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
            break
        fi
        sleep 3s
    done
}

check_obtenant_running() {
    counter=0
    timeout=100  
    OBTENANT_RUNNING='false'
    while true; do
        echo 'check obtenant resource'
        counter=$((counter+1))
        crd_obtenant=`kubectl get obtenant $OBTENANT_NAME -n $NAMESPACE| grep "$OBTENANT_NAME"|awk -F' ' '{print $2}'`
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
	crd_obtenantbackuppolicies_ARCHIVE=`kubectl get obtenantbackuppolicies.oceanbase.oceanbase.com -n $NAMESPACE| grep "$OBTENANT_BACKUPPOLICY_NAME"|awk -F' ' '{print $2}'`
	crd_FULL=`kubectl get obtenantbackups -n $NAMESPACE| grep "full"|awk -F' ' '{print $3}'| head -n 1`
        if [[ $crd_obtenantbackuppolicies_ARCHIVE = "RUNNING" && $crd_FULL = "SUCCESSFUL" ]];then
            echo "backup resource is $crd_obtenantbackuppolicies_ARCHIVE ready"
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
    timeout=150  
    RESTORE_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check RESTORE resource'
	counter=$((counter+1))
	crd_standby_restore=`kubectl get obtenant -n $NAMESPACE | grep "$OBTENANT_STANDBY"|awk -F' ' '{print $2}'`
	crd_restore_obtenant=`kubectl get obtenantrestore -n $NAMESPACE| grep "$OBTENANT_STANDBY"|awk -F' ' '{print $2}'`
	echo "crd_standby_restore $crd_standby_restore  crd_restore_obtenant $crd_restore_obtenant"
        if [[ $crd_standby_restore = "running" && $crd_restore_obtenant = "SUCCESSFUL" ]];then
            RESTORE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "restore resource still not running"
            break
        fi
        sleep 3s
    done
}

check_restore_db_running() {
    counter=0
    timeout=150  
    RESTORE_DB_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check restore db resource'
        counter=$((counter+1))
        tenant_role=`obclient -h $ip -P2881 -A -uroot -p$PASSWORD -Doceanbase -e "select * from DBA_OB_TENANTS;"| grep $OBTENANT_STANDBY|awk -F' ' '{print $15}'`
        if [[ $tenant_role = "PRIMARY" ]];then
            echo "tenant_role $tenant_role"
            RESTORE_DB_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "restore db resource still not running"
            break
        fi
        sleep 3s
    done
}

check_obtenantoperation_running() {
    counter=0
    timeout=10  
    OPERATION_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check obtenantoperation resource'
        counter=$((counter+1))
        crd_obtenantoperation=`kubectl get obtenantoperation -n $NAMESPACE | grep "$OP_ACTIVATE"|awk -F' ' '{print $3}'`
        if [[ $crd_obtenantoperation = "SUCCESSFUL" ]];then
            echo "crd_obtenantoperation $crd_obtenantoperation"
            OPERATION_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "obtenantoperation resource still not running"
            break
        fi
        sleep 2s
    done
}

validate() {
    run "obcluster_template_1-1-1.yaml"
    echo 'do validate'
    check_resource_running
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	run "obtenant_basic_1_1_1.yaml"
	check_obtenant_running
	run "backup_policy_1.yaml"
	check_backup_running
	run "standby_01.yaml"
	check_restore_running
	run "tenant_op_activate.yaml"
	check_obtenantoperation_running	
	check_restore_db_running
	cleanoperation
        if [[ $OPERATION_RUNNING == 'false' || $RESTORE_DB_RUNNING == 'false' ]]; then
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
