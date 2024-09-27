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
    local output_file="case_p4_ResourceModification/env_221_vars.sh"
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
    kubectl delete namespace $NS_DEFAULT $NS_OCEANBASE_TEST
    rm -rf case_p4_ResourceModification/env_221_vars.sh 
}

run() {
    local template_file=$1
    echo $OBCLUSTER_NAME
    envsubst < ./config/2.2.1_test/$template_file | kubectl apply -f -
    envsubst < ./config/2.2.1_test/$template_file 
}

check_ob_image() {
    local obcluster_name=$1
    local namespace=$2
    local cluster_id=$3
    counter=0
    timeout=100  
    OB_IMAGE_NOT_RUNNING='false'
    while true; do
        echo 'check ob image'
        counter=$((counter+1))
        output=$(kubectl describe obcluster $obcluster_name -n $namespace | tail -n 1)
        pod_exists=$(kubectl get pods -n $namespace | grep $obcluster_name-$cluster_id)
	echo $output
	echo $pod_exists
        if  echo "$output" | grep -q "Current version is lower than 4.2.1.4, does not support service mode" && [  -z "$pod_exists" ]; then
	    kubectl describe obcluster $obcluster_name -n $namespace | tail -n 1
            OB_IMAGE_NOT_RUNNING='true'
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

validate() {
    run "obcluster_service_4213.yaml"
    check_ob_image $OBCLUSTER_RECOVERY $NS_DEFAULT $CLUSTER_ID_RECOVERY
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
	echo "case failed"
    else
	echo "case pass"
    fi 
}


#prepare
#export_to_file
validate
cleanup
