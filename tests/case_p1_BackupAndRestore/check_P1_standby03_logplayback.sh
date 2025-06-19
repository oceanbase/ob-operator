#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p1_BackupAndRestore/env_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export OBTENANT_STANDBY_EMPTY=tenantstandyempty$SUFFIX
}

 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    ssh ${USER_NAME}@${NFS_SERVER} "sudo rm -rf ${NFS_BASE_PATH}/*"
}

cleanrestore(){
    kubectl delete obtenant $OBTENANT_STANDBY_EMPTY  -n $NAMESPACE
}

writedata(){
    echo 'write data'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    mysql -h $ip -P2881 -A -uroot@$OBTENANT_NAME -p$PASSWORD  -Dtest -e "create table if not exists demo (id int, value varchar(32));"
    mysql -h $ip -P2881 -A -uroot@$OBTENANT_NAME -p$PASSWORD  -Dtest -e "insert into demo values (1, '321'), (2, '123');"
}

compared_data(){
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    mysql -h $ip -P2881 -A -uroot@$OBTENANT_NAME -p$PASSWORD -Dtest -e "select * from demo;" > result1.txt
    counter=0
    timeout=100  
    COMPARED_DATA='false'
    while true; do
        echo 'compared data'
        counter=$((counter+1))
        mysql -h $ip -P2881 -A -uroot@$OBTENANT_STANDBY_EMPTY -p$PASSWORD -Dtest -e "select * from demo;" > result2.txt
        diff result1.txt result2.txt > /dev/null
        if [ $? -eq 0 ]; then
            cat result1.txt
            cat result2.txt
            rm -rf result1.txt result2.txt
            COMPARED_DATA='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
            break
        fi
        sleep 3s
    done
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
        recovery_window=`mysql -h $ip -uroot@$OBTENANT_NAME -A -P2881 -p$PASSWORD  -Doceanbase -e "select policy_name,recovery_window from DBA_OB_BACKUP_DELETE_POLICY"|grep 8d |awk -F' ' '{print $2}'`
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
            echo "restore db resource still not running"
            break
        fi
        sleep 3s
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
	check_backup_db_running
	run "standby_02.yaml"
	check_restore_running
	writedata
	compared_data
        if [[ $COMPARED_DATA == 'false'  ]]; then
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
