# Shared by the two point-in-time OSS restore scenarios.
oss_test_sql() {
    local tenant=$1 query=$2
    MYSQL_PWD="$PASSWORD" mysql --batch --skip-column-names --connect-timeout=5 \
        -h "$ip" -P 2881 -u "root@$tenant" -Dtest -e "$query"
}

wait_for_oss_restore() {
    local tenant=$1 expected_rows=$2 replay_until=${3:-}
    local cutoff_query cutoff_reached
    local attempt state="" job="" role="" primary_rows restored_rows
    local query='SELECT id, name, age, address FROM students ORDER BY id'
    local all_rows=$'1\tAlice\t20\t123 Main St\n2\tBob\t22\t456 Elm St'
    if [[ -n $replay_until ]]; then
        [[ $tenant =~ ^[A-Za-z0-9_]+$ && $replay_until =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}\ [0-9]{2}:[0-9]{2}:[0-9]{2}$ ]] || return 1
        cutoff_query="SELECT COUNT(*) FROM oceanbase.DBA_OB_TENANTS WHERE TENANT_NAME='$tenant' AND SYNC_SCN=RECOVERY_UNTIL_SCN AND RECOVERY_UNTIL_SCN=TIMESTAMP_TO_SCN('$replay_until')"
    fi
    for ((attempt=0; attempt<150; attempt++)); do
        if state=$(kubectl -n "$NAMESPACE" get obtenant "$tenant" -o jsonpath='{.status.status}') &&
        job=$(kubectl -n "$NAMESPACE" get obtenantrestore "$tenant-restore" -o jsonpath='{.status.status}') &&
        [[ $state == running && $job == SUCCESSFUL ]] &&
        role=$(get_tenant_role "$ip" "$PASSWORD" "$tenant") && [[ $role == STANDBY ]] &&
        primary_rows=$(oss_test_sql "$OBTENANT_NAME" "$query") && [[ $primary_rows == "$all_rows" ]] &&
        restored_rows=$(oss_test_sql "$tenant" "$query") && [[ $restored_rows == "$expected_rows" ]]; then
            if [[ -z $replay_until ]] || { cutoff_reached=$(oss_test_sql sys "$cutoff_query") && [[ $cutoff_reached == 1 ]]; }; then
                return 0
            fi
        fi
        [[ $job == FAILED ]] && break
        sleep 3
    done
    echo "Restore check failed for $tenant (tenant=$state, job=$job, role=$role)" >&2
    return 1
}
