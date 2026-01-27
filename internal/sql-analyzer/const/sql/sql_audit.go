/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package sql

const (
	GetMaxRequestIDByIP             = "SELECT svr_ip, MAX(request_id) as max_request_id FROM gv$ob_sql_audit WHERE tenant_id = ? and svr_port = ? GROUP BY svr_ip"
	CreateSqlAuditTempTableTemplate = `CREATE TEMP TABLE ` + "%s" + ` (
        svr_ip VARCHAR, svr_port BIGINT, tenant_id BIGINT, tenant_name VARCHAR, user_id BIGINT, user_name VARCHAR,
        db_id BIGINT, db_name VARCHAR, sql_id VARCHAR, plan_id BIGINT,
        query_sql TEXT, client_ip VARCHAR, event VARCHAR,
        effective_tenant_id BIGINT, trace_id VARCHAR, sid BIGINT,
        user_client_ip VARCHAR, tx_id VARCHAR,
        executions BIGINT, min_request_time BIGINT, max_request_time BIGINT,
        max_request_id BIGINT, min_request_id BIGINT,
        elapsed_time_sum BIGINT, elapsed_time_max BIGINT, elapsed_time_min BIGINT,
        execute_time_sum BIGINT, execute_time_max BIGINT, execute_time_min BIGINT,
        queue_time_sum BIGINT, queue_time_max BIGINT, queue_time_min BIGINT,
        get_plan_time_sum BIGINT, get_plan_time_max BIGINT, get_plan_time_min BIGINT,
        affected_rows_sum BIGINT, affected_rows_max BIGINT, affected_rows_min BIGINT,
        return_rows_sum BIGINT, return_rows_max BIGINT, return_rows_min BIGINT,
        partition_count_sum BIGINT, partition_count_max BIGINT, partition_count_min BIGINT,
        retry_count_sum BIGINT, retry_count_max BIGINT, retry_count_min BIGINT,
        disk_reads_sum BIGINT, disk_reads_max BIGINT, disk_reads_min BIGINT,
        rpc_count_sum BIGINT, rpc_count_max BIGINT, rpc_count_min BIGINT,
        memstore_read_row_count_sum BIGINT, memstore_read_row_count_max BIGINT, memstore_read_row_count_min BIGINT,
        ssstore_read_row_count_sum BIGINT, ssstore_read_row_count_max BIGINT, ssstore_read_row_count_min BIGINT,
        request_memory_used_sum BIGINT, request_memory_used_max BIGINT, request_memory_used_min BIGINT,
        wait_time_micro_sum BIGINT, wait_time_micro_max BIGINT, wait_time_micro_min BIGINT,
        total_wait_time_micro_sum BIGINT, total_wait_time_micro_max BIGINT, total_wait_time_micro_min BIGINT,
        net_time_sum BIGINT, net_time_max BIGINT, net_time_min BIGINT,
        net_wait_time_sum BIGINT, net_wait_time_max BIGINT, net_wait_time_min BIGINT,
        decode_time_sum BIGINT, decode_time_max BIGINT, decode_time_min BIGINT,
        application_wait_time_sum BIGINT, application_wait_time_max BIGINT, application_wait_time_min BIGINT,
        concurrency_wait_time_sum BIGINT, concurrency_wait_time_max BIGINT, concurrency_wait_time_min BIGINT,
        user_io_wait_time_sum BIGINT, user_io_wait_time_max BIGINT, user_io_wait_time_min BIGINT,
        schedule_time_sum BIGINT, schedule_time_max BIGINT, schedule_time_min BIGINT,
        row_cache_hit_sum BIGINT, row_cache_hit_max BIGINT, row_cache_hit_min BIGINT,
        bloom_filter_cache_hit_sum BIGINT, bloom_filter_cache_hit_max BIGINT, bloom_filter_cache_hit_min BIGINT,
        block_cache_hit_sum BIGINT, block_cache_hit_max BIGINT, block_cache_hit_min BIGINT,
        index_block_cache_hit_sum BIGINT, index_block_cache_hit_max BIGINT, index_block_cache_hit_min BIGINT,
        expected_worker_count_sum BIGINT, expected_worker_count_max BIGINT, expected_worker_count_min BIGINT,
        used_worker_count_sum BIGINT, used_worker_count_max BIGINT, used_worker_count_min BIGINT,
        table_scan_sum BIGINT, table_scan_max BIGINT, table_scan_min BIGINT,
        consistency_level_strong_count BIGINT,
        consistency_level_weak_count BIGINT,
        fail_count_sum BIGINT,
		ret_code_4012_count_sum BIGINT, ret_code_4013_count_sum BIGINT, ret_code_5001_count_sum BIGINT,
		ret_code_5024_count_sum BIGINT, ret_code_5167_count_sum BIGINT, ret_code_5217_count_sum BIGINT,
		ret_code_6002_count_sum BIGINT,
		event_0_wait_time_sum BIGINT, event_1_wait_time_sum BIGINT, event_2_wait_time_sum BIGINT,
		event_3_wait_time_sum BIGINT,
		plan_type_local_count BIGINT, plan_type_remote_count BIGINT, plan_type_distributed_count BIGINT,
		inner_sql_count BIGINT,
		miss_plan_count BIGINT,
		executor_rpc_count BIGINT,
        collect_time TIMESTAMPTZ,
        collect_date DATE
    )`

	GetSqlStatistics = `
		SELECT
			svr_ip, svr_port, tenant_id, tenant_name, user_id, user_name, db_id, db_name, sql_id, plan_id,
			MAX(query_sql) as query_sql, MAX(client_ip) as client_ip, MAX(event) as event, 
			MAX(effective_tenant_id) as effective_tenant_id, MAX(trace_id) as trace_id, MAX(sid) as sid, MAX(user_client_ip) as user_client_ip, MAX(tx_id) as tx_id,
			COUNT(*) as executions, MIN(request_time) as min_request_time, MAX(request_time) as max_request_time,
			MAX(request_id) as max_request_id, MIN(request_id) as min_request_id,
			SUM(elapsed_time) as elapsed_time_sum, MAX(elapsed_time) as elapsed_time_max, MIN(elapsed_time) as elapsed_time_min,
			SUM(execute_time) as execute_time_sum, MAX(execute_time) as execute_time_max, MIN(execute_time) as execute_time_min,
			SUM(queue_time) as queue_time_sum, MAX(queue_time) as queue_time_max, MIN(queue_time) as queue_time_min,
			SUM(get_plan_time) as get_plan_time_sum, MAX(get_plan_time) as get_plan_time_max, MIN(get_plan_time) as get_plan_time_min,
			SUM(affected_rows) as affected_rows_sum, MAX(affected_rows) as affected_rows_max, MIN(affected_rows) as affected_rows_min,
			SUM(return_rows) as return_rows_sum, MAX(return_rows) as return_rows_max, MIN(return_rows) as return_rows_min,
			SUM(partition_cnt) as partition_count_sum, MAX(partition_cnt) as partition_count_max, MIN(partition_cnt) as partition_count_min,
			SUM(retry_cnt) as retry_count_sum, MAX(retry_cnt) as retry_count_max, MIN(retry_cnt) as retry_count_min,
			SUM(disk_reads) as disk_reads_sum, MAX(disk_reads) as disk_reads_max, MIN(disk_reads) as disk_reads_min,
			SUM(rpc_count) as rpc_count_sum, MAX(rpc_count) as rpc_count_max, MIN(rpc_count) as rpc_count_min,
			SUM(memstore_read_row_count) as memstore_read_row_count_sum, MAX(memstore_read_row_count) as memstore_read_row_count_max, MIN(memstore_read_row_count) as memstore_read_row_count_min,
			SUM(ssstore_read_row_count) as ssstore_read_row_count_sum, MAX(ssstore_read_row_count) as ssstore_read_row_count_max, MIN(ssstore_read_row_count) as ssstore_read_row_count_min,
			SUM(request_memory_used) as request_memory_used_sum, MAX(request_memory_used) as request_memory_used_max, MIN(request_memory_used) as request_memory_used_min,
			SUM(wait_time_micro) as wait_time_micro_sum, MAX(wait_time_micro) as wait_time_micro_max, MIN(wait_time_micro) as wait_time_micro_min,
			SUM(total_wait_time_micro) as total_wait_time_micro_sum, MAX(total_wait_time_micro) as total_wait_time_micro_max, MIN(total_wait_time_micro) as total_wait_time_micro_min,
			SUM(net_time) as net_time_sum, MAX(net_time) as net_time_max, MIN(net_time) as net_time_min,
			SUM(net_wait_time) as net_wait_time_sum, MAX(net_wait_time) as net_wait_time_max, MIN(net_wait_time) as net_wait_time_min,
			SUM(decode_time) as decode_time_sum, MAX(decode_time) as decode_time_max, MIN(decode_time) as decode_time_min,
			SUM(application_wait_time) as application_wait_time_sum, MAX(application_wait_time) as application_wait_time_max, MIN(application_wait_time) as application_wait_time_min,
			SUM(concurrency_wait_time) as concurrency_wait_time_sum, MAX(concurrency_wait_time) as concurrency_wait_time_max, MIN(concurrency_wait_time) as concurrency_wait_time_min,
			SUM(user_io_wait_time) as user_io_wait_time_sum, MAX(user_io_wait_time) as user_io_wait_time_max, MIN(user_io_wait_time) as user_io_wait_time_min,
			SUM(schedule_time) as schedule_time_sum, MAX(schedule_time) as schedule_time_max, MIN(schedule_time) as schedule_time_min,
			SUM(row_cache_hit) as row_cache_hit_sum, MAX(row_cache_hit) as row_cache_hit_max, MIN(row_cache_hit) as row_cache_hit_min,
			SUM(bloom_filter_cache_hit) as bloom_filter_cache_hit_sum, MAX(bloom_filter_cache_hit) as bloom_filter_cache_hit_max, MIN(bloom_filter_cache_hit) as bloom_filter_cache_hit_min,
			SUM(block_cache_hit) as block_cache_hit_sum, MAX(block_cache_hit) as block_cache_hit_max, MIN(block_cache_hit) as block_cache_hit_min,
			SUM(index_block_cache_hit) as index_block_cache_hit_sum, MAX(index_block_cache_hit) as index_block_cache_hit_max, MIN(index_block_cache_hit) as index_block_cache_hit_min,
			SUM(expected_worker_count) as expected_worker_count_sum, MAX(expected_worker_count) as expected_worker_count_max, MIN(expected_worker_count) as expected_worker_count_min,
			SUM(used_worker_count) as used_worker_count_sum, MAX(used_worker_count) as used_worker_count_max, MIN(used_worker_count) as used_worker_count_min,
			SUM(table_scan) as table_scan_sum, MAX(table_scan) as table_scan_max, MIN(table_scan) as table_scan_min,
			SUM(CASE WHEN consistency_level = 3 THEN 1 ELSE 0 END) as consistency_level_strong_count,
			SUM(CASE WHEN consistency_level = 2 THEN 1 ELSE 0 END) as consistency_level_weak_count,
			SUM(CASE WHEN ret_code = 0 THEN 0 ELSE 1 END) as fail_count_sum,
			SUM(CASE WHEN ret_code = -4012 THEN 1 ELSE 0 END) as ret_code_4012_count_sum,
			SUM(CASE WHEN ret_code = -4013 THEN 1 ELSE 0 END) as ret_code_4013_count_sum,
			SUM(CASE WHEN ret_code = -5001 THEN 1 ELSE 0 END) as ret_code_5001_count_sum,
			SUM(CASE WHEN ret_code = -5024 THEN 1 ELSE 0 END) as ret_code_5024_count_sum,
			SUM(CASE WHEN ret_code = -5167 THEN 1 ELSE 0 END) as ret_code_5167_count_sum,
			SUM(CASE WHEN ret_code = -5217 THEN 1 ELSE 0 END) as ret_code_5217_count_sum,
			SUM(CASE WHEN ret_code = -6002 THEN 1 ELSE 0 END) as ret_code_6002_count_sum,
			SUM(CASE event WHEN 'system internal wait' THEN wait_time_micro ELSE 0 END) as event_0_wait_time_sum,
			SUM(CASE event WHEN 'mysql response wait client' THEN wait_time_micro ELSE 0 END) as event_1_wait_time_sum,
			SUM(CASE event WHEN 'sync rpc' THEN wait_time_micro ELSE 0 END) as event_2_wait_time_sum,
			SUM(CASE event WHEN 'db file data read' THEN wait_time_micro ELSE 0 END) as event_3_wait_time_sum,
			SUM(CASE plan_type WHEN 1 THEN 1 ELSE 0 END) as plan_type_local_count,
			SUM(CASE plan_type WHEN 2 THEN 1 ELSE 0 END) as plan_type_remote_count,
			SUM(CASE plan_type WHEN 3 THEN 1 ELSE 0 END) as plan_type_distributed_count,
			SUM(CASE is_inner_sql WHEN 1 THEN 1 ELSE 0 END) as inner_sql_count,
			SUM(CASE is_hit_plan WHEN 1 THEN 0 ELSE 1 END) as miss_plan_count,
			SUM(CASE is_executor_rpc WHEN 1 THEN 1 ELSE 0 END) as executor_rpc_count
		FROM gv$ob_sql_audit
		WHERE tenant_id = ? AND svr_ip = ? and svr_port = ? AND request_id > ? and query_sql is not NULL and query_sql <> ''
		GROUP BY
			svr_ip, svr_port, tenant_id, tenant_name, user_id, user_name, db_id, db_name, sql_id, plan_id
	`

	GetMaxRequestIDFromParquet = "SELECT svr_ip, MAX(max_request_id) FROM read_parquet('%s') GROUP BY svr_ip"

	QueryRequestStatisticsTotals = `
		SELECT
			sum(executions),
			sum(fail_count_sum),
			sum(elapsed_time_sum) / sum(executions)
		%s %s`

	QueryRequestStatisticsTrends = `
		SELECT
			strftime(to_timestamp(CAST(max_request_time / 1000000 AS BIGINT)), '%%Y-%%m-%%d') AS day,
			sum(executions),
			sum(elapsed_time_sum) / sum(executions)
		%s %s
		GROUP BY day
		ORDER BY day`

	QueryExecutionTrend = `
		SELECT
			CAST(epoch(time_bucket(INTERVAL %d SECOND, CAST(to_timestamp(CAST(max_request_time / 1000000 AS BIGINT)) AS TIMESTAMP))) AS BIGINT) AS time_bucket,
			sum(plan_type_local_count),
			sum(plan_type_remote_count),
			sum(plan_type_distributed_count)
		FROM read_parquet('%s/*.parquet')
		WHERE
			sql_id = ?
			AND max_request_time >= ?
			AND max_request_time <= ?
		GROUP BY time_bucket
		ORDER BY time_bucket`

	QueryLatencyTrend = `
			SELECT
				CAST(epoch(time_bucket(INTERVAL %d SECOND, CAST(to_timestamp(CAST(max_request_time / 1000000 AS BIGINT)) AS TIMESTAMP))) AS BIGINT) AS time_bucket,
				%s
			FROM read_parquet('%s/*.parquet')
			WHERE
				sql_id = ?
				AND max_request_time >= ?
				AND max_request_time <= ?
			GROUP BY time_bucket
			ORDER BY time_bucket`

	QuerySqlById = `
		SELECT
			query_sql
		FROM
			read_parquet('%s/*.parquet')
		WHERE
			sql_id = ?
			AND max_request_time >= ?
			AND max_request_time <= ?
		LIMIT 1`

	GetTenantIDByName = "SELECT tenant_id FROM __all_tenant WHERE tenant_name = ?"
)
