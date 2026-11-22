#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p4_ResourceModification/env_221_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
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
    
}

export_to_file() {
    local output_file="case_p4_ResourceModification/env_221_vars.sh"
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
    [[ $RESOURCE_RUNNING == true ]]
}
 
check_obtenant_running() {
    local tenant=$1 namespace=$2 expected_role=$3
    local limit=300 observed
    # CREATE STANDBY TENANT has a ten-minute SQL deadline on the operator.
    if [[ $expected_role == STANDBY ]]; then
        limit=600
    fi
    local deadline=$((SECONDS + limit))
    while (( SECONDS < deadline )); do
        if observed=$(kubectl -n "$namespace" get "obtenant/$tenant" \
                -o jsonpath='{.status.status}/{.status.tenantRole}'); then
            if [[ $observed == "running/$expected_role" ]]; then
                echo "$tenant ready as $expected_role"
                return 0
            fi
            if [[ $observed == failed/* ]]; then
                echo "$namespace/$tenant failed: $observed" >&2
                return 1
            fi
        fi
        sleep 3
    done
    echo "$namespace/$tenant did not become running/$expected_role within ${limit}s (last state: ${observed:-unavailable})" >&2
    return 1
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
    [[ $BACKUP_RUNNING == true ]]
}

wait_for_recoverable_backup() {
    local ip ready query
    ip=$(kubectl get pods -n "$NS_DEFAULT" -l "ref-obcluster=$OBCLUSTER_DEFAULT_ONE" \
        --field-selector=status.phase=Running -o jsonpath='{.items[0].status.podIP}') || return 1
    # A completed FULL may still be ahead of the available archive checkpoint.
    query="SELECT EXISTS (
        SELECT 1 FROM CDB_OB_BACKUP_SET_FILES b
        JOIN CDB_OB_ARCHIVELOG a ON a.TENANT_ID = b.TENANT_ID
        JOIN DBA_OB_TENANTS t ON t.TENANT_ID = b.TENANT_ID
        WHERE t.TENANT_NAME = '$TENANT_DEFAULT_ONE' AND b.BACKUP_TYPE = 'FULL'
          AND b.STATUS = 'SUCCESS' AND b.FILE_STATUS = 'AVAILABLE'
          AND a.START_SCN <= b.START_REPLAY_SCN AND a.CHECKPOINT_SCN >= b.MIN_RESTORE_SCN
    )"
    local deadline=$((SECONDS + 600))
    while (( SECONDS < deadline )); do
        if ready=$(MYSQL_PWD="$PASSWORD" mysql --connect-timeout=5 -h "$ip" -P 2881 \
                -u root@sys -NB -Doceanbase -e "$query") && [[ $ready == 1 ]]; then
            return 0
        fi
        sleep 3
    done
    echo "No completed FULL with sufficient archive coverage for $TENANT_DEFAULT_ONE" >&2
    return 1
}

check_restore_policy_running() {
    local tenant=$1 namespace=$2
    kubectl wait -n "$namespace" "obtenant/$tenant" \
        --for=jsonpath='{.status.status}'=running --timeout=15m &&
        kubectl wait -n "$namespace" "obtenantrestore/$tenant-restore" \
            --for=jsonpath='{.status.status}'=SUCCESSFUL --timeout=1m
}

validate() {
    run "obcluster_ns_test.yaml" || return 1
    run "obcluster_ns_default.yaml" || return 1
    run "obcluster_ns_default2.yaml" || return 1
    check_resource_running "$OBCLUSTER_OCEANBASE_TEST" "$NS_OCEANBASE_TEST" "$CLUSTER_ID_TEST" || return 1
    check_resource_running "$OBCLUSTER_DEFAULT_ONE" "$NS_DEFAULT" "$CLUSTER_ID_DEFAULT_ONE" || return 1
    check_resource_running "$OBCLUSTER_DEFAULT_TWO" "$NS_DEFAULT" "$CLUSTER_ID_DEFAULT_TWO" || return 1
    run "tenant_ns_default.yaml" || return 1
    check_obtenant_running "$TENANT_DEFAULT_ONE" "$NS_DEFAULT" PRIMARY || return 1
    run "tenant_ns_default_2.yaml" || return 1
    run "tenant_ns_test_empty.yaml" || return 1
    run "tenant_default_backup_policy.yaml" || return 1
    check_obtenant_running "$TENANT_DEFAULT_TWO" "$NS_DEFAULT" STANDBY || return 1
    check_obtenant_running "$TENANT_OCEANBASE_EMPTY" "$NS_OCEANBASE_TEST" STANDBY || return 1
    check_backup_policy_running "$OBTENANT_DEFAULT_BACK_POLICY" "$NS_DEFAULT" || return 1
    wait_for_recoverable_backup || return 1
    run "tenant_ns_test_restore.yaml" || return 1
    check_restore_policy_running "$TENANT_RESTORE" "$NS_OCEANBASE_TEST"
}

prepare
export_to_file
if validate; then
    echo "case passed"
    result=0
else
    echo "case failed"
    result=1
fi
cleanup
exit "$result"
