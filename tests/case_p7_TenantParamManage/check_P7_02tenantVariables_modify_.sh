#!/bin/bash
 
# load all the parameters in setup.sh
source setup.sh
source util.sh
source env.sh
#source case_p6/env_vars.sh

# prepare create related resources to create obcluster
prepare() {
    export PASSWORD=$(generate_random_str)
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
    echo $PASSWORD
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
	    echo "case passed"
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "obtenant resource still not running"
	    echo "case failed"
            break
        fi
        sleep 3s
    done
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

modifyparameters() {
    local value="$1"
    kubectl patch obtenant $OBTENANT_NAME -n $NAMESPACE --type json -p "[{"op": "add", "path": "/spec/variables", "value": [{"name": "autocommit", "value": \"$value\"}]}]"
}

removeparameters() {
    local value="$1"
    kubectl patch obtenant $OBTENANT_NAME -n $NAMESPACE --type json -p "[{"op": "remove", "path": "/spec/variables", "value": [{"name": \"$value\"}]}]"
}


check_crd_obtenantvariables_running() {
    local expected_value="$1"
    counter=0
    timeout=10
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
	counter=$((counter+1))
        crd_variables=`kubectl get obtenantvariables -A | grep autocommit |awk -F' ' '{print $6}'| awk 'NR==1'`
        if [[ $crd_variables = "$expected_value" ]];then
            RESOURCE_RUNNING='true'
	    echo "autocommit"
	    echo $crd_variables
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
            break
        fi
        sleep 3s
    done
}

check_crd_obtenantvariables_empty() {
    counter=0
    timeout=10
    RESOURCE_RUNNING='false'
    while true; do
        echo 'check resource'
	counter=$((counter+1))
        crd_obparameter=`kubectl get obtenantvariables -A | grep autocommit |awk -F' ' '{print $6}'| awk 'NR==1'`
        if [[ $crd_obparameter = "" ]];then
            RESOURCE_RUNNING='true'
            echo "ob_vector_memory_limit_percentage"
            echo $crd_obparameter
            break
        fi
        if [ $counter -eq $timeout ]; then
            echo "resource still not running"
            break
        fi
        sleep 3s
    done
}

validate_obparameter() {
    echo 'do validate'
    local expected_value="$1" 
    check_crd_obtenantvariables_running $expected_value
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
	cleanup
	prepare
    else
	echo "case passed"
    fi
}

validate_obparameter_empty() {
    echo 'do validate'
    check_crd_obtenantvariables_empty
    if [[ $RESOURCE_RUNNING == 'false' ]]; then
        echo "case failed"
        cleanup
        prepare
    else
        echo "case passed"
    fi
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

#cleanup 
prepare
export_to_file
run "obcluster_template_1-1-1.yaml"
validate 
run "obtenant_basic_1_1_1.yaml"
check_obtenant_running
modifyparameters 1
validate_obparameter 1
modifyparameters 0
validate_obparameter 0
removeparameters "autocommit"
validate_obparameter_empty
cleanup
