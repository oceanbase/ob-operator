#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
source case_p4_v220_and_v211/env_221_vars.sh
 
# prepare create related resources to create obcluster
prepare() {
    export OPERATOR_NS="oceanbase-system"
}

# clean up delete everything by deleting the entire namespace
cleanup() {
    kubectl delete namespace $OPERATOR_NS 
}

run() {
    local template_file=$1
    envsubst < ./config/2.2.1_test/$template_file | kubectl apply -f -
#    envsubst < ./config/2.2.1_test/$template_file 
}


operate() {
    local value=$1
    kubectl patch -n $OPERATOR_NS deployment oceanbase-controller-manager -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","env":[{"name":"OB_OPERATOR_TASK_POOLSIZE","value":"$value"}]}]}}}}'
}

check_modify_value_running() {
    local value=$1
    counter=0
    timeout=100  
    MODIFY_VALUE_RUNNING='false'
    while true; do
        echo 'check modify value  resource'
        counter=$((counter+1))
 	if output=$(kubectl get pods -n oceanbase-system | grep 'oceanbase-controller-manager' | awk '{print $1}' | xargs -I{} kubectl logs -n oceanbase-system {} | head -n 40 | grep manager) && echo "$output" | grep -q '"Debug":false,"PoolSize":9999'; then
            MODIFY_VALUE_RUNNING='true'
            break
        fi
        if [ $counter -eq $timeout ]; then
            break
        fi
        sleep 3s
    done
}

check_nodify_success(){
    kubectl patch -n oceanbase-system deployment oceanbase-controller-manager -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","env":[{"name":"OB_OPERATOR_TASK_POOLSIZE","value":"9999"}]}]}}}}'
    sleep 15s
    if output=$(kubectl get pods -n oceanbase-system | grep 'oceanbase-controller-manager' | awk '{print $1}' | xargs -I{} kubectl logs -n oceanbase-system {} | head -n 40 | grep manager) && echo "$output" | grep -q '"Debug":false,"PoolSize":9999'; then
        echo "pass"
    else
        echo "case-fail"
    fi

    kubectl get pods -n oceanbase-system | grep 'oceanbase-controller-manager' | awk '{print $1}' | xargs -I{} kubectl logs -n oceanbase-system {} | head -n 40 | grep manager
    kubectl patch -n oceanbase-system deployment oceanbase-controller-manager -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","env":[{"name":"OB_OPERATOR_TASK_POOLSIZE","value":"10000"}]}]}}}}'
    sleep 15s
    if output=$(kubectl get pods -n oceanbase-system | grep 'oceanbase-controller-manager' | awk '{print $1}' | xargs -I{} kubectl logs -n oceanbase-system {} | head -n 40 | grep manager) && echo "$output" | grep -q '"Debug":false,"PoolSize":10000'; then
        echo "case-pass"
    else
        echo "case-fail"
fi
    kubectl get pods -n oceanbase-system | grep 'oceanbase-controller-manager' | awk '{print $1}' | xargs -I{} kubectl logs -n oceanbase-system {} | head -n 40 | grep manager

}

validate() {
    operate 9999
    check_modify_value_running 9999
    if [[ $MODIFY_VALUE_RUNNING == 'false' ]]; then
	echo "case failed"
    else
	operate 10000
        check_modify_value_running 10000
        if [[ $MODIFY_VALUE_RUNNING == 'false' ]]; then
            echo "case failed"
        else
            echo "case passed" 
        fi
    fi 
}

check_nodify_success
