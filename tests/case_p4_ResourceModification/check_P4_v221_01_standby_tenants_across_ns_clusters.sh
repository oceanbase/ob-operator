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
    export OBCLUSTER_OCEANBASE_TEST=obcluster-oceanbase-test-$SUFFIX
    export NS_OCEANBASE_TEST=$NAMESPACE_PREFIX-ocenbasetest-$SUFFIX
    export CLUSTER_ID_TEST=$(( (RANDOM % 9000) + 1000 ))
    export OB_ROOT_SECRET=ob-root-sc-$SUFFIX
    export SA_DEFAULT=sa-default-$SUFFIX
    kubectl create namespace $NS_OCEANBASE_TEST
    kubectl create sa $SA_DEFAULT -n $NS_OCEANBASE_TEST
    create_pass_secret $NS_OCEANBASE_TEST $OB_ROOT_SECRET $PASSWORD

    export OBCLUSTER_DEFAULT_ONE=obcluster-default-1-$SUFFIX
    export NS_DEFAULT=$NAMESPACE_PREFIX-default-$SUFFIX
    export CLUSTER_ID_DEFAULT_ONE=$(( (RANDOM % 9000) + 1000 ))
    kubectl create namespace $NS_DEFAULT
    kubectl create sa $SA_DEFAULT -n $NS_DEFAULT
    create_pass_secret $NS_DEFAULT $OB_ROOT_SECRET $PASSWORD

    export OBCLUSTER_DEFAULT_TWO=obcluster-default-2-$SUFFIX
    export CLUSTER_ID_DEFAULT_TWO=$(( (RANDOM % 9000) + 1000 ))
    
    export TENANT_DEFAULT_ONE=tenantdefault1$SUFFIX
    export BACKUP_ROOT_SECRET=backup-root-serect
    export BACKUP_STANDBY_SECRET=backup-standby-serect
    create_pass_secret $NS_DEFAULT $BACKUP_ROOT_SECRET $PASSWORD
    create_pass_secret $NS_DEFAULT $BACKUP_STANDBY_SECRET $PASSWORD
    
    export TENANT_DEFAULT_TWO=tenantdefault2$SUFFIX
    
    export TENANT_OCEANBASE_EMPTY=tenantempty$SUFFIX
    create_pass_secret $NS_OCEANBASE_TEST $BACKUP_ROOT_SECRET $PASSWORD
    create_pass_secret $NS_OCEANBASE_TEST $BACKUP_STANDBY_SECRET $PASSWORD

    export OBTENANT_DEFAULT_BACK_POLICY=obtenantdefaultbackuppolicy$SUFFIX    
    export OSS_BACPUP_PATH=backup-path-$SUFFIX
    export OSS_ARCHIVE_PATH=archive-path-$SUFFIX
    export OSS_ACCESS=oss-access-$SUFFIX
    create_pass_oss_secret $NS_OCEANBASE_TEST $OSS_ACCESS $OSS_AK $OSS_SK
    create_pass_oss_secret $NS_DEFAULT $OSS_ACCESS $OSS_AK $OSS_SK
     
    export TENANT_RESTORE=tenantrestore$SUFFIX
    
    echo $PASSWORD
}

export_to_file() {
    local output_file="case_p4_v220_and_v211/env_221_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export OBCLUSTER_OCEANBASE_TEST="$OBCLUSTER_OCEANBASE_TEST"
export NS_OCEANBASE_TEST="$NS_OCEANBASE_TEST"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export SA_DEFAULT="$SA_DEFAULT"
export OBCLUSTER_DEFAULT_ONE="$OBCLUSTER_DEFAULT_ONE"
export NS_DEFAULT="$NS_DEFAULT"
export OBCLUSTER_DEFAULT_TWO="$OBCLUSTER_DEFAULT_TWO"
export TENANT_DEFAULT_ONE="$TENANT_DEFAULT_ONE"
export BACKUP_ROOT_SECRET="$BACKUP_ROOT_SECRET"
export BACKUP_STANDBY_SECRET="$BACKUP_STANDBY_SECRET"
export TENANT_DEFAULT_TWO="$TENANT_DEFAULT_TWO"
export TENANT_OCEANBASE_EMPTY="$TENANT_OCEANBASE_EMPTY"
export OBTENANT_DEFAULT_BACK_POLICY="$OBTENANT_DEFAULT_BACK_POLICY"
export OSS_BACPUP_PATH="$OSS_BACPUP_PATH"
export OSS_ARCHIVE_PATH="$OSS_ARCHIVE_PATH"
export OSS_ACCESS="$OSS_ACCESS"
export TENANT_RESTORE="$TENANT_RESTORE"
export CLUSTER_ID_TEST="$CLUSTER_ID_TEST"
export CLUSTER_ID_DEFAULT_ONE="$CLUSTER_ID_DEFAULT_ONE"
export CLUSTER_ID_DEFAULT_TWO="$CLUSTER_ID_DEFAULT_TWO"
EOF
    echo "Environment variables have been exported to $output_file"
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NS_OCEANBASE_TEST $NS_DEFAULT 
}

run() {
    local template_file=$1
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

check_backup_policy_running() {
    local backuppolicy_name=$1
    local namespace=$2
    counter=0
    timeout=100  
    BACKUP_RUNNING='false'
    while true; do
        echo 'check backup resource'
        counter=$((counter+1))
        crd_obtenantbackuppolicies_ARCHIVE=`kubectl get obtenantbackuppolicies.oceanbase.oceanbase.com -n $namespace| grep "$backuppolicy_name"|awk -F' ' '{print $2}'`
        crd_FULL=`kubectl get obtenantbackups -n $namespace| grep "full"|awk -F' ' '{print $3}'| head -n 1`
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

check_restore_policy_running() {
    local restore_policy_name=$1
    local namespace=$2
    counter=0
    timeout=120  
    RESTORE_RUNNING='false'
    while true; do
        echo 'check restore resource'
	counter=$((counter+1))
        crd_primary_restore=`kubectl get obtenant -n $namespace| grep "$restore_policy_name"|awk -F' ' '{print $2}'`
        crd_primary_obtenant=`kubectl get obtenantrestore -n $namespace| grep "$restore_policy_name"|awk -F' ' '{print $2}'`
        echo "crd_primary_restore $crd_primary_restore  crd_primary_obtenant $crd_primary_obtenant"
        if [[ $crd_primary_restore = "running" && $crd_primary_obtenant = "SUCCESSFUL" ]];then
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
    run "obcluster_ns_test.yaml"
    run "obcluster_ns_default.yaml"
    run "obcluster_ns_default2.yaml"
    echo 'do validate'
    check_resource_running $OBCLUSTER_OCEANBASE_TEST $NS_OCEANBASE_TEST $CLUSTER_ID_TEST
    check_resource_running $OBCLUSTER_DEFAULT_ONE $NS_DEFAULT $CLUSTER_ID_DEFAULT_ONE
    check_resource_running $OBCLUSTER_DEFAULT_TWO $NS_DEFAULT $CLUSTER_ID_DEFAULT_TWO
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        run "tenant_ns_default.yaml"
	check_obtenant_running $TENANT_DEFAULT_ONE $NS_DEFAULT
  	run "tenant_ns_default_2.yaml"
	run "tenant_ns_test_empty.yaml"
	run "tenant_default_backup_policy.yaml"
	check_obtenant_running $TENANT_DEFAULT_TWO $NS_DEFAULT
	check_obtenant_running $TENANT_OCEANBASE_EMPTY $NS_OCEANBASE_TEST
	check_backup_policy_running $OBTENANT_DEFAULT_BACK_POLICY $NS_DEFAULT
	sleep 60s
	run "tenant_ns_test_restore.yaml"
	check_restore_policy_running $TENANT_RESTORE $NS_OCEANBASE_TEST
	check_obtenant_running $TENANT_RESTORE $NS_OCEANBASE_TEST
        if [[ $RESTORE_RUNNING == 'false' ]]; then
            echo "case failed"
        else
            echo "case passed"
        fi
    fi
}

prepare
export_to_file
validate
cleanup
