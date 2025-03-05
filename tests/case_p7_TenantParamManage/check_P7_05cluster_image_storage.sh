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
	IMAGE=`kubectl get obcluster -A -o wide |grep $OBCLUSTER_NAME |awk -F' ' '{print $6}'| awk 'NR==1'`
	STORAGE=`kubectl get obcluster -A -o wide |grep $OBCLUSTER_NAME |awk -F' ' '{print $7}'| awk 'NR==1'`
	echo $IMAGE $STORAGE
        if [[ $server_1_zone1 == "ACTIVE" && $server_1_zone2 == "ACTIVE" && $server_1_zone3 == "ACTIVE" && "$IMAGE" == "oceanbase/oceanbase-cloud-native:4.2.1.6-106000012024042515" && "$STORAGE" == '{"dataStorage":{"size":"60Gi","storageClass":"local-path"},"logStorage":{"size":"20Gi","storageClass":"local-path"},"redoLogStorage":{"size":"60Gi","storageClass":"local-path"}}' ]]
        then
            echo "server_1_zone1 $server_1_zone1"
            echo "server_1_zone2 $server_1_zone2"
            echo "server_1_zone3 $server_1_zone3"
	    echo $IMAGE
	    echo $STORAGE
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
cleanup
