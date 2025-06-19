#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p1_BackupAndRestore/env_210_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export OBTENANT_STANDBY_OSS_REPLAY=restoresstandbyossrp$SUFFIX
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}

run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.1.0_test/$template_file | kubectl apply -f -
#    envsubst < ./config/2.1.0_test/$template_file 
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
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
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
            echo "restore resource still not running"
            break
        fi
        sleep 3s
    done
}

check_restore_db_running() {
    counter=0
    timeout=120  
    RESTORE_DB_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check restore db resource'
        counter=$((counter+1))
        tenant_role=`mysql -h $ip -P2881 -A -uroot -p$PASSWORD -Doceanbase -e "select * from DBA_OB_TENANTS;"| grep $OBTENANT_STANDBY_OSS|awk -F' ' '{print $15}'`
        if [[ $tenant_role = "STANDBY" ]];then
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

check_restore_data() {
    counter=0
    timeout=100  
    RESTORE_DATA_SAME='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check dbdata resource'
        counter=$((counter+1))
	mysql  -uroot@$OBTENANT_NAME -h $ip -P2881 -p$PASSWORD -Dtest -e " select * from students;" > result1.txt
	mysql  -uroot@$OBTENANT_STANDBY_OSS -h $ip -P2881 -p$PASSWORD -Dtest -e " select * from students;" > result2.txt
	diff result1.txt result2.txt > /dev/null
        if [ $? -eq 0 ]; then
	    cat result1.txt && cat result2.txt
	    rm -rf result1.txt result2.txt
            echo "tenant_role $tenant_role"
            RESTORE_DATA_SAME='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "backup dbdata resource still not running"
	    cat result1.txt && cat result2.txt 
	    rm -rf result1.txt result2.txt
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
	run "backup_policy_oss.yaml"
        check_backup_running
	run "tenant_restore_until_replay.yaml"
	check_restore_running
	check_restore_db_running
	check_restore_data
	echo $OSS_REPLOY_LOG_UNTIL
        if [[ $RESTORE_DATA_SAME == 'false' ]]; then
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
