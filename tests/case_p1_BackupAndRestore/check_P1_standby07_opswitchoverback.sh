#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p1_BackupAndRestore/env_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export OBTENANT_STANDBY_EMPTY=tenantstandyempty$SUFFIX
    export OP_SWITCHOVER_BACK=op-switchoverback$SUFFIX
}

 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    ssh ${USER_NAME}@${NFS_SERVER} "sudo rm -rf ${NFS_BASE_PATH}/*"
    rm -rf case_p1_BackupAndRestore/env_vars.sh
}

cleanrestore(){
    kubectl delete obtenant $OBTENANT_STANDBY_EMPTY  -n $NAMESPACE
}
cleanobtenantoperation() {
    kubectl delete obtenantoperation $OP_SWITCHOVER_BACK -n $NAMESPACE
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
        if [[ $server_1_zone1 == "ACTIVE" ]]
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

check_backup_db_running() {
    counter=0
    timeout=100  
    BACKUP_DB_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check backup resource'
        counter=$((counter+1))
        recovery_window=`obclient -h $ip -uroot@$OBTENANT_NAME -A -P2881 -p$PASSWORD  -Doceanbase -e "select policy_name,recovery_window from DBA_OB_BACKUP_DELETE_POLICY"|grep 8d |awk -F' ' '{print $2}'`
        if [[ $recovery_window = "8d" ]];then
            echo "recovery_window $recovery_window"
            BACKUP_DB_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "backup db resource still not running"
            break
        fi
        sleep 3s
    done
}

check_restore_running() {
    counter=0
    timeout=100  
    RESTORE_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check RESTORE resource'
	counter=$((counter+1))
	crd_standby_restore=`kubectl get obtenant -n $NAMESPACE | grep "$OBTENANT_STANDBY_EMPTY"|awk -F' ' '{print $2}'`
	echo "crd_standby_restore $crd_standby_restore  "
        if [[ $crd_standby_restore = "running"  ]];then
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

check_obtenantoperation_running() {
    counter=0
    timeout=10  
    OPERATION_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check obtenantoperation resource'
        counter=$((counter+1))
        crd_obtenantoperation=`kubectl get obtenantoperation -n $NAMESPACE | grep "$OP_SWITCHOVER_BACK"|awk -F' ' '{print $3}'`
	standby_empty_tenantrole=`kubectl get obtenant -n $NAMESPACE| grep "$OBTENANT_STANDBY_EMPTY" | awk -F' ' '{print $4}' | head -n 1`
	primary_tenantrole=`kubectl get obtenant -n $NAMESPACE| grep "$OBTENANT_NAME" | awk -F' ' '{print $4}' | head -n 1`	
	echo "crd_obtenantoperation $crd_obtenantoperation standby_empty_tenantrole is $standby_empty_tenantrole primary_tenantrole is $primary_tenantrole "
        if [[ $crd_obtenantoperation = "SUCCESSFUL" && $standby_empty_tenantrole = "STANDBY" && $primary_tenantrole = "PRIMARY" ]];then
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
	run "standby_02.yaml"
	check_restore_running
	run "tenant_op_switchover_back.yaml"
	check_obtenantoperation_running
	cleanobtenantoperation
        if [[ $OPERATION_RUNNING = 'false' ]]; then
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
cleanup
