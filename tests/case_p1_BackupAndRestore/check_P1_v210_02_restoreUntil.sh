#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
#source case_p1_BackupAndRestore/env_210_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export PASSWORD_ENCRYPTION=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=test$SUFFIX
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
    timeout=150  
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
    timeout=150  
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
    timeout=150  
    RESTORE_DB_RUNNING='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check restore db resource'
        counter=$((counter+1))
        tenant_role=`obclient -h $ip -P2881 -A -uroot -p$PASSWORD -Doceanbase -e "select * from DBA_OB_TENANTS;"| grep $OBTENANT_STANDBY_OSS|awk -F' ' '{print $15}'`
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
    timeout=150  
    RESTORE_DATA_SAME='false'
    ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
    while true; do
        echo 'check dbdata resource'
        counter=$((counter+1))
	obclient  -uroot@$OBTENANT_NAME -h $ip -P2881 -p$PASSWORD -Dtest -e " select * from students;" > result1.txt
	obclient  -uroot@$OBTENANT_STANDBY_OSS -h $ip -P2881 -p$PASSWORD -Dtest -e " select * from students;" > result2.txt
	diff result1.txt result2.txt > /dev/null
        if [ $? -ne 0 ]; then
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


export_to_file() {
    local output_file="case_p1_BackupAndRestore/env_210_vars.sh"
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
EOF
    echo "Environment variables have been exported to $output_file"
}

insert_data() {
   echo "insert first data"
   ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
   obclient  -uroot@$OBTENANT_NAME -h$ip -P2881 -p$PASSWORD -Dtest -e " DROP TABLE IF EXISTS students;"
   obclient  -uroot@$OBTENANT_NAME -h$ip -P2881 -p$PASSWORD -Dtest -e " CREATE TABLE students ( id INT PRIMARY KEY, name VARCHAR(50), age INT, address VARCHAR(100));"
   obclient  -uroot@$OBTENANT_NAME -h$ip -P2881 -p$PASSWORD -Dtest -e " INSERT INTO students (id, name, age, address) VALUES (1, 'Alice', 20, '123 Main St');"
}

insert_data_second() {
   echo "insert second data" 
   ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
   obclient  -uroot@$OBTENANT_NAME -h$ip -P2881 -p$PASSWORD -Dtest -e " INSERT INTO students (id, name, age, address) VALUES (2, 'Bob', 22, '456 Elm St');"
}

export_localtime() {
    local template_var=\$1
    local datetime=$(date "+%Y-%m-%d %H:%M:%S")
    eval "export $template_var=\"$datetime\""
}



validate() {
    run "obcluster_template_1-1-1.yaml"
    echo 'do validate'
    check_resource_running
    sleep 20s
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster
	run "obtenant_basic_1_1_1.yaml"
	check_obtenant_running
	run "backup_policy_oss.yaml"
        check_backup_running
	sleep 100s
	insert_data
	sleep 30s
	export_localtime "OSS_UNTIL_TIMESTAMP"
	echo $OSS_UNTIL_TIMESTAMP
	sleep 30s
	insert_data_second
	sleep 100s
	run "tenant_restore_until.yaml"
	check_restore_running
	check_restore_db_running
	check_restore_data
	export_localtime "OSS_REPLOY_LOG_UNTIL"
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
export_to_file
validate
export_to_file
