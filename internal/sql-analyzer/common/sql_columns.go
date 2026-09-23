/*
Copyright (c) 2026 OceanBase
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

import httperr "github.com/oceanbase/ob-operator/pkg/errors"

// FixedDimensions are always projected and grouped in SQL statistics queries.
var FixedDimensions = map[string]struct{}{
	"tenant_name": {},
	"user_name":   {},
	"db_name":     {},
	"sql_id":      {},
	"plan_id":     {},
}

// Dimensions are projected using MAX when requested by the caller.
var Dimensions = map[string]struct{}{
	"svr_ip":              {},
	"svr_port":            {},
	"tenant_id":           {},
	"user_id":             {},
	"db_id":               {},
	"query_sql":           {},
	"client_ip":           {},
	"event":               {},
	"effective_tenant_id": {},
	"trace_id":            {},
	"sid":                 {},
	"user_client_ip":      {},
	"tx_id":               {},
	"sub_plan_count":      {},
	"last_fail_info":      {},
	"cause_type":          {},
}

// ValidateSortColumn accepts an optional, exact catalog key, never a SQL
// expression. Reuse the projection catalogs rather than trusting outputColumns.
func ValidateSortColumn(column string) error {
	if column == "" {
		return nil
	}
	if _, ok := ColumnAggregations[column]; ok {
		return nil
	}
	if _, ok := FixedDimensions[column]; ok {
		return nil
	}
	if _, ok := Dimensions[column]; ok {
		return nil
	}
	return httperr.NewBadRequest("invalid SQL sort column")
}
