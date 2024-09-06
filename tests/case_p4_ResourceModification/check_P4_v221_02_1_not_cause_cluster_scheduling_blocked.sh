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
    export OBCLUSTER_RECOVERY=obcluster-recovery-$SUFFIX
    export NS_DEFAULT=$NAMESPACE_PREFIX-default-$SUFFIX
    export CLUSTER_ID_RECOVERY=$(( (RANDOM % 9000) + 1000 ))
    export OB_ROOT_SECRET=ob-root-sc-$SUFFIX
    export SA_DEFAULT=sa-default-$SUFFIX

    export OBCLUSTER_DEFAULT_TWO=obcluster-default-two-$SUFFIX
    export CLUSTER_ID_DEFAULT_TWO=$(( (RANDOM % 9000) + 1000 ))
    kubectl create namespace $NS_DEFAULT
    kubectl create sa $SA_DEFAULT -n $NS_DEFAULT
    create_pass_secret $NS_DEFAULT $OB_ROOT_SECRET $PASSWORD
}

export_to_file() {
    local output_file="case_p4_v220_and_v211/env_221_vars.sh"
    cat <<EOF > "$output_file"
export PASSWORD="$PASSWORD"
export SUFFIX="$SUFFIX"
export OBCLUSTER_RECOVERY="$OBCLUSTER_RECOVERY"
export NS_DEFAULT="$NS_DEFAULT"
export CLUSTER_ID_RECOVERY="$CLUSTER_ID_RECOVERY"
export OB_ROOT_SECRET="$OB_ROOT_SECRET"
export SA_DEFAULT="$SA_DEFAULT"
export OBCLUSTER_DEFAULT_TWO="$OBCLUSTER_DEFAULT_TWO"
export CLUSTER_ID_DEFAULT_TWO="$CLUSTER_ID_DEFAULT_TWO"
EOF
    echo "Environment variables have been exported to $output_file"
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NS_DEFAULT 
}

run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
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
 	    exit 1
            break
        fi
        sleep 3s
    done
}
 
kill_two_pod() {
    pod1=`kubectl get pods -n $NS_DEFAULT -l ref-obcluster=$OBCLUSTER_RECOVERY | grep  $OBCLUSTER_RECOVERY-$CLUSTER_ID_RECOVERY-zone2 | awk '{print $1}'`
    pod2=`kubectl get pods -n $NS_DEFAULT -l ref-obcluster=$OBCLUSTER_RECOVERY | grep  $OBCLUSTER_RECOVERY-$CLUSTER_ID_RECOVERY-zone3 | awk '{print $1}'`
    kubectl delete pods $pod1 $pod2 -n $NS_DEFAULT
}

check_both_active() {
    counter=0
    timeout=100  
    OBCLUSTER_BOTH_RUNNING='false'
    while true; do
        echo 'check resource'
        counter=$((counter+1))
        obcluster2_crd=`kubectl get obcluster $OBCLUSTER_DEFAULT_TWO -n $NS_DEFAULT | grep -w $OBCLUSTER_DEFAULT_TWO | awk '{print $2}'`
        obcluster3_crd=`kubectl get obcluster $OBCLUSTER_RECOVERY -n $NS_DEFAULT | grep -w $OBCLUSTER_RECOVERY | awk '{print $2}'`
	echo $OBCLUSTER_DEFAULT_TWO $obcluster2_crd
	echo $OBCLUSTER_RECOVERY  $obcluster3_crd
	if [[ $obcluster2_crd = "running" && $obcluster3_crd = "running"   ]]; then
            OBCLUSTER_BOTH_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            break
        fi
        sleep 3s
    done

}

validate() {
    run "obcluster_service_3_nodes.yaml"
    check_resource_running  $OBCLUSTER_RECOVERY $NS_DEFAULT $CLUSTER_ID_RECOVERY
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
	echo "case failed"
    else
        kill_two_pod
        run "obcluster_ns_default2.yaml"
        check_both_active
        echo $OBCLUSTER_BOTH_RUNNING
        if [[ $OBCLUSTER_BOTH_RUNNING == 'false' ]]; then
            echo "case failed"
        else
            echo "case passed" 
        fi
    fi 
}

date
prepare
export_to_file
validate
cleanup
date
