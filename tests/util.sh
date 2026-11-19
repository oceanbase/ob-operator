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

create_pass_cos_secret() {
    local namespace=$1
    local secret_name=$2
    local accessId=$3
    local accessKey=$4
    local appid=$5
    kubectl create secret generic $secret_name -n $namespace --from-literal=accessId=$accessId --from-literal=accessKey=$accessKey --from-literal=appId=$appid
}

create_pass_s3_secret() {
    local namespace=$1
    local secret_name=$2
    local accessId=$3
    local accessKey=$4
    local s3region=$5
    kubectl create secret generic $secret_name -n $namespace --from-literal=accessId=$accessId --from-literal=accessKey=$accessKey --from-literal=s3Region=$s3region
}

create_pass_obs_secret() {
    local namespace=$1
    local secret_name=$2
    local accessId=$3
    local accessKey=$4
    kubectl create secret generic $secret_name -n $namespace --from-literal=accessId=$accessId --from-literal=accessKey=$accessKey 
}

create_pass_gcs_secret() {
    local namespace=$1
    local secret_name=$2
    local accessId=$3
    local accessKey=$4
    kubectl create secret generic $secret_name -n $namespace --from-literal=accessId=$accessId --from-literal=accessKey=$accessKey
}

clean_nfs_test_directory() {
    local path=${NFS_BASE_PATH:-} normalized quoted
    if [[ -z ${USER_NAME:-} || -z ${NFS_SERVER:-} || $path != /* || $path == *[[:cntrl:]]* ]]; then
        echo "NFS cleanup requires a user, server and absolute test directory" >&2
        return 1
    fi
    normalized=$(realpath -ms -- "$path") || return 1
    if [[ $normalized == / || /"${path#/}"/ == */../* ]]; then
        echo "NFS cleanup requires a non-root test directory without parent traversal" >&2
        return 1
    fi
    quoted=${path//\'/\'\\\'\'}
    # Resolve symlinks on the NFS server before expanding the wildcard there.
    ssh -- "$USER_NAME@$NFS_SERVER" "cd -P -- '$quoted' && dir=\$(pwd -P) && [ -n \"\$dir\" ] && [ \"\$dir\" != / ] && sudo rm -rf -- ./*"
}

get_tenant_role() {
    local host=$1 password=$2 tenant=$3
    MYSQL_PWD="$password" mysql -h "$host" -P 2881 -uroot -Doceanbase --batch --skip-column-names \
        -e 'SELECT TENANT_NAME, TENANT_ROLE FROM DBA_OB_TENANTS' |
        awk -F '\t' -v tenant="$tenant" '$1 == tenant { print $2 }'
}
