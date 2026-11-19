#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p0_ClusterAndTenant/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    kubectl create namespace $NAMESPACE
    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
}
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/tenantManage/$template_file | kubectl apply -f -
}
 
check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone1" | awk -v line="1" 'NR==line{print \$2}'`
        pod_1_zone2=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone2" | awk -v line="1" 'NR==line{print \$2}'`
        pod_1_zone3=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone3" | awk -v line="1" 'NR==line{print \$2}'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "1/1" && $pod_1_zone2 = "1/1" && $pod_1_zone3 = "1/1" && -n "$ip" && $crd_obcluster = "status:running" ]];then
            echo "pod_1_zone1 is $pod_1_zone1 ready"
            echo "pod_1_zone2 is $pod_1_zone2 ready"
            echo "pod_1_zone3 is $pod_1_zone3 ready"
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
    while true; do
        echo 'check ob'
        counter=$((counter+1))
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1| awk -v line="1" 'NR==line{print \$11}'`
        server_1_zone2=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone2| awk -v line="1" 'NR==line{print \$11}'`
        server_1_zone3=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone3| awk -v line="1" 'NR==line{print \$11}'`
        if [[ $server_1_zone1 == "ACTIVE" && $server_1_zone2 == "ACTIVE" && $server_1_zone3 == "ACTIVE"  ]]
        then
            echo "server_1_zone1 $server_1_zone1"
            echo "server_1_zone2 $server_1_zone2"
            echo "server_1_zone3 $server_1_zone3"
            OBSERVER_ACTIVE='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
	    echo "case failed"
            break
        fi
        sleep 3s
    done
}
 
wait_for_zone2_pool() {
    local active=$1 bit=0 ip observed state record until
    [[ $OBTENANT_NAME =~ ^[[:alnum:]]+$ ]] || return 1
    [[ $active == true ]] && bit=1
    ip=$(kubectl -n "$NAMESPACE" get pods -l "ref-obcluster=$OBCLUSTER_NAME" -o jsonpath='{.items[0].status.podIP}') || return 1
    until=$((SECONDS + 600))
    while ((SECONDS < until)); do
        observed=$(kubectl -n "$NAMESPACE" get "obtenant/$OBTENANT_NAME" -o json) || return 1
        state=$(jq -r '.status.status' <<< "$observed")
        if [[ $state == failed ]]; then echo "Tenant failed waiting for zone2 active=$active" >&2; return 1; fi
        if [[ $state == running ]] && jq -e --argjson active "$active" '
          any(.status.resourcePool[]?; .zoneList == "zone2" and .type.isActive == $active)
        ' <<< "$observed" >/dev/null; then
            record=$(MYSQL_PWD="$PASSWORD" mysql --batch --skip-column-names --connect-timeout=3 -h "$ip" -P2881 -uroot -Doceanbase -e "
              SELECT T.STATUS, T.PREVIOUS_LOCALITY IS NULL, LOCATE('@zone2,', CONCAT(T.LOCALITY, ','))>0,
                     R.RESOURCE_POOL_ID, GROUP_CONCAT(U.UNIT_ID ORDER BY U.UNIT_ID)
              FROM DBA_OB_TENANTS T JOIN DBA_OB_RESOURCE_POOLS R ON R.TENANT_ID=T.TENANT_ID
              JOIN DBA_OB_UNITS U ON U.RESOURCE_POOL_ID=R.RESOURCE_POOL_ID
              WHERE T.TENANT_NAME='$OBTENANT_NAME' AND R.ZONE_LIST='zone2'
              GROUP BY T.TENANT_ID,T.STATUS,T.PREVIOUS_LOCALITY,T.LOCALITY,R.RESOURCE_POOL_ID") || return 1
            if [[ $record == $'NORMAL\t1\t'"$bit"$'\t'* ]]; then
                pool_identity=$(cut -f4- <<< "$record")
                echo "zone2 active=$active with database pool and units $pool_identity"
                return 0
            fi
        fi
        sleep 3
    done
    echo "Timed out waiting for the existing zone2 pool and active=$active (CR $state, SQL ${record:-pending})" >&2
    return 1
}

validate() {
    echo 'do validate'
    check_resource_running 
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster 
        if [[ $OBSERVER_ACTIVE == 'false' ]]; then
            echo "case failed"
	    cleanup
	    prepare
        else
            echo "case passed"
        fi
    fi
}
 
run "obcluster_template_1-1-1.yaml"
validate 
run "obtenant_basic_1_1_1.yaml" || exit 1
wait_for_zone2_pool true || { echo "case failed"; exit 1; }
active_identity=$pool_identity
run "obtenant_basic_1_1_1_reserve.yaml" || exit 1
wait_for_zone2_pool false || { echo "case failed"; exit 1; }
if [[ $pool_identity != "$active_identity" ]]; then
    echo "case failed: reserved pool or unit was replaced"
    exit 1
fi
echo "case passed"
