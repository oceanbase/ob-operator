#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p2_v211/env_vars.sh

#prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
    export OB_CLUSTER_ID=$(( (RANDOM % 10000) + 1000 ))
    export SUFFIX=$(generate_random_str | tr '[:upper:]' '[:lower:]')
    export NAMESPACE=$NAMESPACE_PREFIX-$SUFFIX
    export OBCLUSTER_NAME=test$SUFFIX
    export OB_ROOT_SECRET=sc-root-$SUFFIX
    export OBTENANT_NAME=tenant$SUFFIX
    kubectl create namespace $NAMESPACE
    create_pass_secret $NAMESPACE $OB_ROOT_SECRET $PASSWORD
    echo $PASSWORD
}
 
# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $NAMESPACE
}

run() {
    local output_yaml=$1
    envsubst < ./config/2.1.1_test/$output_yaml | kubectl apply -f -  2>&1
    
}
 
check_webhook() {
    output1=$(run "case4_cluster_wrong_memory.yaml")
    output2=$(run "case4_cluster_small_disk.yaml")
    output3=$(run "case4_cluster_no_secrets.yaml")
    output4=$(run "case4_cluster_no_storageclass.yaml")
    output5=$(run "case4_tenant_with_wrong_cluster.yaml")
    #output6=$(run "case4_tenant_with_wrong_zone.yaml")
    output7=$(run "case4_tenant_with_wrong_name.yaml")
    echo $output1
    echo $output2
    echo $output3
    echo $output4
    echo $output5
    WEBHOOK_ACTIVE='false'
    if  echo "$output1" | grep -q "Memory limit exceeds observer's resource"  &&
        echo "$output2" | grep -q "The minimum size of data storage should be larger than 3 times of memory limit" &&
        echo "$output3" | grep -q "Given root credential ob-user-root not found"  &&
        echo "$output4" | grep -q "storageClass local-path-2 not found" &&
        echo "$output5" | grep -q "Given cluster not found"  &&
   #    echo "$output6" | grep -q ""  &&
        echo "$output7" | grep -q "which should start with character or underscore and contain character, digit and underscore only"; then
        WEBHOOK_ACTIVE='true'
        
    fi

}

validate() {
    check_webhook
    if [[ $WEBHOOK_ACTIVE == 'false' ]]; then
        echo "case failed"
    else
	echo "case passed"
    fi
}

prepare
validate
cleanup
