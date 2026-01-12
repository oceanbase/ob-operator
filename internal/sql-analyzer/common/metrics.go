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

package common

import "fmt"

// ColumnAggregations defines how to aggregate each metric column.
var ColumnAggregations = map[string]string{
	"executions":                     "SUM",
	"min_request_time":               "MIN",
	"max_request_time":               "MAX",
	"max_request_id":                 "MAX",
	"min_request_id":                 "MIN",
	"elapsed_time":                   "AVG",
	"elapsed_time_sum":               "SUM",
	"elapsed_time_max":               "MAX",
	"elapsed_time_min":               "MIN",
	"execute_time":                   "AVG",
	"execute_time_sum":               "SUM",
	"execute_time_max":               "MAX",
	"execute_time_min":               "MIN",
	"queue_time":                     "AVG",
	"queue_time_sum":                 "SUM",
	"queue_time_max":                 "MAX",
	"queue_time_min":                 "MIN",
	"get_plan_time":                  "AVG",
	"get_plan_time_sum":              "SUM",
	"get_plan_time_max":              "MAX",
	"get_plan_time_min":              "MIN",
	"affected_rows":                  "AVG",
	"affected_rows_sum":              "SUM",
	"affected_rows_max":              "MAX",
	"affected_rows_min":              "MIN",
	"return_rows":                    "AVG",
	"return_rows_sum":                "SUM",
	"return_rows_max":                "MAX",
	"return_rows_min":                "MIN",
	"partition_count":                "AVG",
	"partition_count_sum":            "SUM",
	"partition_count_max":            "MAX",
	"partition_count_min":            "MIN",
	"retry_count":                    "AVG",
	"retry_count_sum":                "SUM",
	"retry_count_max":                "MAX",
	"retry_count_min":                "MIN",
	"disk_reads":                     "AVG",
	"disk_reads_sum":                 "SUM",
	"disk_reads_max":                 "MAX",
	"disk_reads_min":                 "MIN",
	"rpc_count":                      "AVG",
	"rpc_count_sum":                  "SUM",
	"rpc_count_max":                  "MAX",
	"rpc_count_min":                  "MIN",
	"memstore_read_row_count":        "AVG",
	"memstore_read_row_count_sum":    "SUM",
	"memstore_read_row_count_max":    "MAX",
	"memstore_read_row_count_min":    "MIN",
	"ssstore_read_row_count":         "AVG",
	"ssstore_read_row_count_sum":     "SUM",
	"ssstore_read_row_count_max":     "MAX",
	"ssstore_read_row_count_min":     "MIN",
	"request_memory_used":            "AVG",
	"request_memory_used_sum":        "SUM",
	"request_memory_used_max":        "MAX",
	"request_memory_used_min":        "MIN",
	"wait_time_micro":                "AVG",
	"wait_time_micro_sum":            "SUM",
	"wait_time_micro_max":            "MAX",
	"wait_time_micro_min":            "MIN",
	"total_wait_time_micro":          "AVG",
	"total_wait_time_micro_sum":      "SUM",
	"total_wait_time_micro_max":      "MAX",
	"total_wait_time_micro_min":      "MIN",
	"net_time":                       "AVG",
	"net_time_sum":                   "SUM",
	"net_time_max":                   "MAX",
	"net_time_min":                   "MIN",
	"net_wait_time":                  "AVG",
	"net_wait_time_sum":              "SUM",
	"net_wait_time_max":              "MAX",
	"net_wait_time_min":              "MIN",
	"decode_time":                    "AVG",
	"decode_time_sum":                "SUM",
	"decode_time_max":                "MAX",
	"decode_time_min":                "MIN",
	"application_wait_time":          "AVG",
	"application_wait_time_sum":      "SUM",
	"application_wait_time_max":      "MAX",
	"application_wait_time_min":      "MIN",
	"concurrency_wait_time":          "AVG",
	"concurrency_wait_time_sum":      "SUM",
	"concurrency_wait_time_max":      "MAX",
	"concurrency_wait_time_min":      "MIN",
	"user_io_wait_time":              "AVG",
	"user_io_wait_time_sum":          "SUM",
	"user_io_wait_time_max":          "MAX",
	"user_io_wait_time_min":          "MIN",
	"schedule_time":                  "AVG",
	"schedule_time_sum":              "SUM",
	"schedule_time_max":              "MAX",
	"schedule_time_min":              "MIN",
	"row_cache_hit":                  "AVG",
	"row_cache_hit_sum":              "SUM",
	"row_cache_hit_max":              "MAX",
	"row_cache_hit_min":              "MIN",
	"bloom_filter_cache_hit":         "AVG",
	"bloom_filter_cache_hit_sum":     "SUM",
	"bloom_filter_cache_hit_max":     "MAX",
	"bloom_filter_cache_hit_min":     "MIN",
	"block_cache_hit":                "AVG",
	"block_cache_hit_sum":            "SUM",
	"block_cache_hit_max":            "MAX",
	"block_cache_hit_min":            "MIN",
	"index_block_cache_hit":          "AVG",
	"index_block_cache_hit_sum":      "SUM",
	"index_block_cache_hit_max":      "MAX",
	"index_block_cache_hit_min":      "MIN",
	"expected_worker_count":          "AVG",
	"expected_worker_count_sum":      "SUM",
	"expected_worker_count_max":      "MAX",
	"expected_worker_count_min":      "MIN",
	"used_worker_count":              "AVG",
	"used_worker_count_sum":          "SUM",
	"used_worker_count_max":          "MAX",
	"used_worker_count_min":          "MIN",
	"table_scan":                     "AVG",
	"table_scan_sum":                 "SUM",
	"table_scan_max":                 "MAX",
	"table_scan_min":                 "MIN",
	"consistency_level_strong_count": "SUM",
	"consistency_level_weak_count":   "SUM",
	"cpu_time":                       "AVG",
	"cpu_time_sum":                   "SUM",
	"cpu_time_max":                   "MAX",
	"cpu_time_min":                   "MIN",
	"fail_count_sum":                 "SUM",
	"ret_code_4012_count_sum":        "SUM",
	"ret_code_4013_count_sum":        "SUM",
	"ret_code_5001_count_sum":        "SUM",
	"ret_code_5024_count_sum":        "SUM",
	"ret_code_5167_count_sum":        "SUM",
	"ret_code_5217_count_sum":        "SUM",
	"ret_code_6002_count_sum":        "SUM",
	"event_0_wait_time_sum":          "SUM",
	"event_1_wait_time_sum":          "SUM",
	"event_2_wait_time_sum":          "SUM",
	"event_3_wait_time_sum":          "SUM",
	"plan_type_local_count":          "SUM",
	"plan_type_remote_count":         "SUM",
	"plan_type_distributed_count":    "SUM",
	"inner_sql_count":                "SUM",
	"miss_plan_count":                "SUM",
	"executor_rpc_count":             "SUM",
}

var TimeMetrics = map[string]struct{}{
	"elapsed_time":            {},
	"execute_time":            {},
	"queue_time":              {},
	"get_plan_time":           {},
	"wait_time_micro":         {},
	"total_wait_time_micro":   {},
	"net_time":                {},
	"net_wait_time":           {},
	"decode_time":             {},
	"application_wait_time":   {},
	"concurrency_wait_time":   {},
	"user_io_wait_time":       {},
	"schedule_time":           {},
	"event_0_wait_time_sum":   {},
	"event_1_wait_time_sum":   {},
	"event_2_wait_time_sum":   {},
	"event_3_wait_time_sum":   {},
}

// BuildMetricExpression constructs the SQL expression for a given metric column.
func BuildMetricExpression(col string) string {
	if agg, isMetric := ColumnAggregations[col]; isMetric {
		columnExpr := fmt.Sprintf("%s(%s)", agg, col)
		if agg == "AVG" {
			// For average, we calculate it manually using SUM / SUM(executions)
			// assuming the column name without suffix is backed by a _sum column
			// But wait, the map key "elapsed_time" maps to "AVG".
			// We need to know the backing column name.
			// Convention: metric "foo" uses backing column "foo_sum" for AVG calculation.
			columnExpr = fmt.Sprintf("SUM(%s_sum) / NULLIF(SUM(executions), 0)", col)
		}

		if _, isTime := TimeMetrics[col]; isTime {
			// Convert microseconds to milliseconds if needed?
			// The original code in business/sql_stats.go divided by 1000.
			// "The data in parquet is stored as microseconds"
			// And we want milliseconds.
			columnExpr = fmt.Sprintf("(%s) / 1000", columnExpr)
		}
		return columnExpr
	}
	return ""
}
