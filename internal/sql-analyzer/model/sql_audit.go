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

package model

type SqlAudit struct {
	// Grouping Keys
	SvrIP      string `db:"svr_ip"`
	SvrPort    int64  `db:"svr_port"`
	TenantId   uint64 `db:"tenant_id"`
	TenantName string `db:"tenant_name"`
	UserId     int64  `db:"user_id"`
	UserName   string `db:"user_name"`
	DBId       uint64 `db:"db_id"`
	DBName     string `db:"db_name"`
	SqlId      string `db:"sql_id"`
	PlanId     int64  `db:"plan_id"`

	// Aggregated String/Identifier Values
	QuerySql          string `db:"query_sql"`
	ClientIp          string `db:"client_ip"`
	Event             string `db:"event"`
	FormatSqlId       string `db:"format_sql_id"`
	EffectiveTenantId uint64 `db:"effective_tenant_id"`
	TraceId           string `db:"trace_id"`
	Sid               uint64 `db:"sid"`
	UserClientIp      string `db:"user_client_ip"`
	TxId              string `db:"tx_id"`

	// Aggregated Numeric Values
	Executions     int64  `db:"executions"` // COUNT(*)
	MinRequestTime int64  `db:"min_request_time"`
	MaxRequestTime int64  `db:"max_request_time"`
	MaxRequestId   uint64 `db:"max_request_id"` // Crucial for progress tracking
	MinRequestId   uint64 `db:"min_request_id"`

	ElapsedTimeSum int64 `db:"elapsed_time_sum"`
	ElapsedTimeMax int64 `db:"elapsed_time_max"`
	ElapsedTimeMin int64 `db:"elapsed_time_min"`

	ExecuteTimeSum int64 `db:"execute_time_sum"`
	ExecuteTimeMax int64 `db:"execute_time_max"`
	ExecuteTimeMin int64 `db:"execute_time_min"`

	QueueTimeSum int64 `db:"queue_time_sum"`
	QueueTimeMax int64 `db:"queue_time_max"`
	QueueTimeMin int64 `db:"queue_time_min"`

	GetPlanTimeSum int64 `db:"get_plan_time_sum"`
	GetPlanTimeMax int64 `db:"get_plan_time_max"`
	GetPlanTimeMin int64 `db:"get_plan_time_min"`

	AffectedRowsSum int64 `db:"affected_rows_sum"`
	AffectedRowsMax int64 `db:"affected_rows_max"`
	AffectedRowsMin int64 `db:"affected_rows_min"`

	ReturnRowsSum int64 `db:"return_rows_sum"`
	ReturnRowsMax int64 `db:"return_rows_max"`
	ReturnRowsMin int64 `db:"return_rows_min"`

	PartitionCountSum int64 `db:"partition_count_sum"`
	PartitionCountMax int64 `db:"partition_count_max"`
	PartitionCountMin int64 `db:"partition_count_min"`

	RetryCountSum int64 `db:"retry_count_sum"`
	RetryCountMax int64 `db:"retry_count_max"`
	RetryCountMin int64 `db:"retry_count_min"`

	DiskReadsSum int64 `db:"disk_reads_sum"`
	DiskReadsMax int64 `db:"disk_reads_max"`
	DiskReadsMin int64 `db:"disk_reads_min"`

	RpcCountSum int64 `db:"rpc_count_sum"`
	RpcCountMax int64 `db:"rpc_count_max"`
	RpcCountMin int64 `db:"rpc_count_min"`

	MemstoreReadRowCountSum int64 `db:"memstore_read_row_count_sum"`
	MemstoreReadRowCountMax int64 `db:"memstore_read_row_count_max"`
	MemstoreReadRowCountMin int64 `db:"memstore_read_row_count_min"`

	SSStoreReadRowCountSum int64 `db:"ssstore_read_row_count_sum"`
	SSStoreReadRowCountMax int64 `db:"ssstore_read_row_count_max"`
	SSStoreReadRowCountMin int64 `db:"ssstore_read_row_count_min"`

	RequestMemoryUsedSum int64 `db:"request_memory_used_sum"`
	RequestMemoryUsedMax int64 `db:"request_memory_used_max"`
	RequestMemoryUsedMin int64 `db:"request_memory_used_min"`

	WaitTimeMicroSum int64 `db:"wait_time_micro_sum"`
	WaitTimeMicroMax int64 `db:"wait_time_micro_max"`
	WaitTimeMicroMin int64 `db:"wait_time_micro_min"`

	TotalWaitTimeMicroSum int64 `db:"total_wait_time_micro_sum"`
	TotalWaitTimeMicroMax int64 `db:"total_wait_time_micro_max"`
	TotalWaitTimeMicroMin int64 `db:"total_wait_time_micro_min"`

	NetTimeSum int64 `db:"net_time_sum"`
	NetTimeMax int64 `db:"net_time_max"`
	NetTimeMin int64 `db:"net_time_min"`

	NetWaitTimeSum int64 `db:"net_wait_time_sum"`
	NetWaitTimeMax int64 `db:"net_wait_time_max"`
	NetWaitTimeMin int64 `db:"net_wait_time_min"`

	DecodeTimeSum int64 `db:"decode_time_sum"`
	DecodeTimeMax int64 `db:"decode_time_max"`
	DecodeTimeMin int64 `db:"decode_time_min"`

	ApplicationWaitTimeSum int64 `db:"application_wait_time_sum"`
	ApplicationWaitTimeMax int64 `db:"application_wait_time_max"`
	ApplicationWaitTimeMin int64 `db:"application_wait_time_min"`

	ConcurrencyWaitTimeSum int64 `db:"concurrency_wait_time_sum"`
	ConcurrencyWaitTimeMax int64 `db:"concurrency_wait_time_max"`
	ConcurrencyWaitTimeMin int64 `db:"concurrency_wait_time_min"`

	UserIoWaitTimeSum int64 `db:"user_io_wait_time_sum"`
	UserIoWaitTimeMax int64 `db:"user_io_wait_time_max"`
	UserIoWaitTimeMin int64 `db:"user_io_wait_time_min"`

	ScheduleTimeSum int64 `db:"schedule_time_sum"`
	ScheduleTimeMax int64 `db:"schedule_time_max"`
	ScheduleTimeMin int64 `db:"schedule_time_min"`

	RowCacheHitSum int64 `db:"row_cache_hit_sum"`
	RowCacheHitMax int64 `db:"row_cache_hit_max"`
	RowCacheHitMin int64 `db:"row_cache_hit_min"`

	BloomFilterCacheHitSum int64 `db:"bloom_filter_cache_hit_sum"`
	BloomFilterCacheHitMax int64 `db:"bloom_filter_cache_hit_max"`
	BloomFilterCacheHitMin int64 `db:"bloom_filter_cache_hit_min"`

	BlockCacheHitSum int64 `db:"block_cache_hit_sum"`
	BlockCacheHitMax int64 `db:"block_cache_hit_max"`
	BlockCacheHitMin int64 `db:"block_cache_hit_min"`

	IndexBlockCacheHitSum int64 `db:"index_block_cache_hit_sum"`
	IndexBlockCacheHitMax int64 `db:"index_block_cache_hit_max"`
	IndexBlockCacheHitMin int64 `db:"index_block_cache_hit_min"`

	ExpectedWorkerCountSum int64 `db:"expected_worker_count_sum"`
	ExpectedWorkerCountMax int64 `db:"expected_worker_count_max"`
	ExpectedWorkerCountMin int64 `db:"expected_worker_count_min"`

	UsedWorkerCountSum int64 `db:"used_worker_count_sum"`
	UsedWorkerCountMax int64 `db:"used_worker_count_max"`
	UsedWorkerCountMin int64 `db:"used_worker_count_min"`

	TableScanSum int64 `db:"table_scan_sum"`
	TableScanMax int64 `db:"table_scan_max"`
	TableScanMin int64 `db:"table_scan_min"`

	ConsistencyLevelStrongCount int64 `db:"consistency_level_strong_count"`
	ConsistencyLevelWeakCount   int64 `db:"consistency_level_weak_count"`

	FailCountSum int64 `db:"fail_count_sum"`

	RetCode4012CountSum int64 `db:"ret_code_4012_count_sum"`
	RetCode4013CountSum int64 `db:"ret_code_4013_count_sum"`
	RetCode5001CountSum int64 `db:"ret_code_5001_count_sum"`
	RetCode5024CountSum int64 `db:"ret_code_5024_count_sum"`
	RetCode5167CountSum int64 `db:"ret_code_5167_count_sum"`
	RetCode5217CountSum int64 `db:"ret_code_5217_count_sum"`
	RetCode6002CountSum int64 `db:"ret_code_6002_count_sum"`

	Event0WaitTimeSum int64 `db:"event_0_wait_time_sum"`
	Event1WaitTimeSum int64 `db:"event_1_wait_time_sum"`
	Event2WaitTimeSum int64 `db:"event_2_wait_time_sum"`
	Event3WaitTimeSum int64 `db:"event_3_wait_time_sum"`

	PlanTypeLocalCount       int64 `db:"plan_type_local_count"`
	PlanTypeRemoteCount      int64 `db:"plan_type_remote_count"`
	PlanTypeDistributedCount int64 `db:"plan_type_distributed_count"`
	InnerSqlCount            int64 `db:"inner_sql_count"`
	MissPlanCount            int64 `db:"miss_plan_count"`
	ExecutorRpcCount         int64 `db:"executor_rpc_count"`
}
