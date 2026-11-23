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

wait_for_recoverable_nfs_backup() {
    local namespace=$1 cluster=$2 tenant=$3 password=$4 data_path=$5 archive_path=$6
    local ip expected eligible deadline=$((SECONDS + 600))
    ip=$(kubectl -n "$namespace" get pods -l "ref-obcluster=$cluster" \
        --field-selector=status.phase=Running -o jsonpath='{.items[0].status.podIP}') || return 1
    expected=$(printf '%s\tfile:///ob-backup/%s\tfile:///ob-backup/%s' "$tenant" "$data_path" "$archive_path")
    # FULL can finish before archived logs cover its earliest restore SCN.
    local query="SELECT DISTINCT t.TENANT_NAME, b.PATH, a.PATH
        FROM CDB_OB_BACKUP_SET_FILES b
        JOIN CDB_OB_ARCHIVELOG a ON a.TENANT_ID = b.TENANT_ID
        JOIN DBA_OB_TENANTS t ON t.TENANT_ID = b.TENANT_ID
        WHERE b.BACKUP_TYPE = 'FULL' AND b.STATUS = 'SUCCESS' AND b.FILE_STATUS = 'AVAILABLE'
          AND b.PATH LIKE 'file://%' AND a.PATH LIKE 'file://%'
          AND b.BACKUP_SET_ID = (SELECT MAX(s.BACKUP_SET_ID) FROM CDB_OB_BACKUP_SET_FILES s
              WHERE s.TENANT_ID = b.TENANT_ID AND s.PATH = b.PATH AND s.BACKUP_TYPE = 'FULL'
                AND s.STATUS = 'SUCCESS' AND s.FILE_STATUS = 'AVAILABLE')
          AND a.START_SCN <= b.START_REPLAY_SCN AND a.CHECKPOINT_SCN >= b.MIN_RESTORE_SCN"
    echo "Waiting for a recoverable NFS backup of $tenant"
    while (( SECONDS < deadline )); do
        eligible=$(MYSQL_PWD="$password" mysql --connect-timeout=5 -h "$ip" -P 2881 -u root@sys \
            --batch --raw --skip-column-names -Doceanbase -e "$query") || return 1
        if grep -Fqx -- "$expected" <<< "$eligible"; then
            echo "Archived logs cover a completed NFS backup of $tenant"
            return 0
        fi
        sleep 3
    done
    echo "No recoverable FULL for $tenant at $data_path with archive $archive_path" >&2
    return 1
}

get_tenant_role() {
    local host=$1 password=$2 tenant=$3
    MYSQL_PWD="$password" mysql -h "$host" -P 2881 -uroot -Doceanbase --batch --skip-column-names \
        -e 'SELECT TENANT_NAME, TENANT_ROLE FROM DBA_OB_TENANTS' |
        awk -F '\t' -v tenant="$tenant" '$1 == tenant { print $2 }'
}

check_obagent_metrics() {
    local ns=$1 cluster=$2 mode=$3 count=$4 entry pod ip expected url metrics matched attempt
    local -a pods
    for ((attempt=0; attempt<60; attempt++)); do
        mapfile -t pods < <(kubectl -n "$ns" get pods -l "ref-obcluster=$cluster" -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.podIP}{"\n"}{end}')
        matched=0
        for entry in "${pods[@]}"; do
            read -r pod ip <<< "$entry"
            [[ -n $pod && -n $ip ]] || continue
            expected=127.0.0.1
            if [[ $mode == service ]]; then
                expected=$(kubectl -n "$ns" get service "$pod" -o jsonpath='{.spec.clusterIP}') || continue
                [[ -n $expected ]] || continue
            fi
            url=$ip
            [[ $ip == *:* ]] && url="[$ip]"
            metrics=$(curl -fsS --max-time 5 "http://$url:8088/metrics/ob/basic") || continue
            if grep '^ob_active_session_num{' <<< "$metrics" | grep -F "svr_ip=\"$expected\"" >/dev/null; then
                ((matched+=1))
            fi
        done
        if (( matched == count && ${#pods[@]} == count )); then
            echo "$matched OBAgent endpoints reported the expected observer IPs"
            return 0
        fi
        sleep 3
    done
    return 1
}
