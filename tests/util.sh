#!/bin/bash
 
generate_random_str() {
    openssl rand -base64 32 | tr -dc '[:alnum:]' | fold -w 6 | head -n 1
}
 
create_pass_secret() {
    local namespace=$1
    local secret_name=$2
    local password=$3
    kubectl create secret generic $secret_name -n $namespace --from-literal=password=$password
}

create_pass_oss_secret() {
    local namespace=$1
    local secret_name=$2
    local accessId=$3
    local accessKey=$4
    kubectl create secret generic $secret_name -n $namespace --from-literal=accessId=$accessId --from-literal=accessKey=$accessKey
}

