# Shared by the two point-in-time OSS restore scenarios.
oss_test_sql() {
    local tenant=$1 query=$2
    MYSQL_PWD="$PASSWORD" mysql --batch --skip-column-names --connect-timeout=5 \
        -h "$ip" -P 2881 -u "root@$tenant" -Dtest -e "$query"
}

wait_for_oss_restore() {
    local tenant=$1 expected_rows=$2
    local attempt state="" job="" role="" primary_rows restored_rows
    local query='SELECT id, name, age, address FROM students ORDER BY id'
    local all_rows=$'1\tAlice\t20\t123 Main St\n2\tBob\t22\t456 Elm St'
    for ((attempt=0; attempt<150; attempt++)); do
        state=$(kubectl -n "$NAMESPACE" get obtenant "$tenant" -o jsonpath='{.status.status}') &&
        job=$(kubectl -n "$NAMESPACE" get obtenantrestore "$tenant-restore" -o jsonpath='{.status.status}') &&
        [[ $state == running && $job == SUCCESSFUL ]] &&
        role=$(get_tenant_role "$ip" "$PASSWORD" "$tenant") && [[ $role == STANDBY ]] &&
        primary_rows=$(oss_test_sql "$OBTENANT_NAME" "$query") && [[ $primary_rows == "$all_rows" ]] &&
        restored_rows=$(oss_test_sql "$tenant" "$query") && [[ $restored_rows == "$expected_rows" ]] && return 0
        [[ $job == FAILED ]] && break
        sleep 3
    done
    echo "Restore check failed for $tenant (tenant=$state, job=$job, role=$role)" >&2
    return 1
}
