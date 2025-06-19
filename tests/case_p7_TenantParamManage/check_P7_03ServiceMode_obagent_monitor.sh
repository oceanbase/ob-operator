#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
#source case_p4_ResourceModification/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=testservice$SUFFIX
    export OB_ROOT_SECRET=sc-root-$SUFFIX
    export OBTENANT_NAME=tenant$SUFFIX
    kubectl create namespace $NAMESPACE
    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
    echo $PASSWORD
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
    rm -rf case_p4_ResourceModification/env_vars.sh
}

cleanup_obcluster() {
    kubectl delete obcluster $OBCLUSTER_NAME -n  $NAMESPACE
}
 
run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.3.1_test/$template_file | kubectl apply -f -
}
 
check_resource_running() {
    counter=0
    timeout=100  
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        pod_1_zone1=`kubectl get pod -o wide -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone1" | awk -v line="1" 'NR==line{print \$2}'`
        ip=`kubectl get pod  -o wide -n $NAMESPACE | grep $OBCLUSTER_NAME-1-zone1 |awk -F' ' '{print $6}'| awk 'NR==1'`
        crd_obcluster=`kubectl get obcluster $OBCLUSTER_NAME -n $NAMESPACE  -o yaml| grep "status: running" | tail -n 1| sed 's/ //g'`
        if [[ $pod_1_zone1 = "2/2" && -n "$ip" && $crd_obcluster = "status:running" ]];then
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
        server_1_zone1=`mysql -uroot -h $ip -P 2881 -Doceanbase -p$PASSWORD -e 'select * from __all_server;'|grep zone1| awk -v line="1" 'NR==line{print \$11}'`
        if [[ $server_1_zone1 == "ACTIVE"  ]]
        then
            echo "server_1_zone1 $server_1_zone1"
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

check_service_agent() {
    counter=0
    timeout=100
    SERVICE_AGENT='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
	obagent_ip=`kubectl get svc -n $NAMESPACE | grep svc-monitor-$OBCLUSTER_NAME |awk -F' ' '{print $3}'| awk 'NR==1'`
	pod_svc1=`kubectl get svc -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone1" |awk -F' ' '{print $3}'| awk 'NR==1'`
	pod_svc2=`kubectl get svc -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone2" |awk -F' ' '{print $3}'| awk 'NR==1'`
	pod_svc3=`kubectl get svc -n $NAMESPACE | grep "$OBCLUSTER_NAME-1-zone3" |awk -F' ' '{print $3}'| awk 'NR==1'`
	obagent_svr_ip=$(curl -L "http://$obagent_ip:8088/metrics/ob/basic" 2>/dev/null | grep 'ob_active_session_num' | awk -F'svr_ip="' '{print $2}' | awk -F'"' '{print $1}')
	# 清理变量中的不可见字符
        obagent_svr_ip=$(echo "$obagent_svr_ip" | tr -d '[:space:]')
        pod_svc1=$(echo "$pod_svc1" | tr -d '[:space:]')
        pod_svc2=$(echo "$pod_svc2" | tr -d '[:space:]')
        pod_svc3=$(echo "$pod_svc3" | tr -d '[:space:]')

	echo "obagent_ip" $obagent_ip
	echo "pod_svc1" $pod_svc1
	echo "pod_svc2" $pod_svc2
	echo "pod_svc3" $pod_svc3
	echo "obagent_svr_ip" $obagent_svr_ip

	if [[ "$obagent_svr_ip" == "$pod_svc1" || "$obagent_svr_ip" == "$pod_svc2" || "$obagent_svr_ip" == "$pod_svc3" ]]; then
            SERVICE_AGENT='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "obagent_svr_ip not in pod_svc"
            break
        fi
        sleep 3s
    done
}


export_to_file() {
    local output_file="case_p7_TenantParamManage/env_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export NAMESPACE="$NAMESPACE"
export OBCLUSTER_NAME="$OBCLUSTER_NAME"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export OBTENANT_NAME="$OBTENANT_NAME"
EOF
    echo "Environment variables have been exported to $output_file"
}

validate() {
    run "obcluster_service.yaml"
    echo 'do validate'
    check_resource_running 
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
    else
        check_in_obcluster 
	check_service_agent
        if [[ $SERVICE_AGENT == 'false' ]]; then
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
#cleanup_obcluster
cleanup
