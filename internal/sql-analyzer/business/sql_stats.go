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

package business

import (
	"fmt"
	"math/big"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/common"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	"github.com/sirupsen/logrus"
)

var fixedDimensions = map[string]any{
	"tenant_name":   struct{}{},
	"user_name":     struct{}{},
	"db_name":       struct{}{},
	"sql_id":        struct{}{},
	"plan_id":       struct{}{},
	"format_sql_id": struct{}{},
}

var dimensions = map[string]any{
	"svr_ip":              struct{}{},
	"svr_port":            struct{}{},
	"tenant_id":           struct{}{},
	"user_id":             struct{}{},
	"db_id":               struct{}{},
	"query_sql":           struct{}{},
	"client_ip":           struct{}{},
	"event":               struct{}{},
	"effective_tenant_id": struct{}{},
	"trace_id":            struct{}{},
	"sid":                 struct{}{},
	"user_client_ip":      struct{}{},
	"tx_id":               struct{}{},
	"sub_plan_count":      struct{}{},
	"last_fail_info":      struct{}{},
	"cause_type":          struct{}{},
}

type SqlStatsService struct {
	Store  *store.SqlAuditStore
	Config *config.Config
	Logger *logrus.Logger
}

func NewSqlStatsService(store *store.SqlAuditStore, conf *config.Config, logger *logrus.Logger) *SqlStatsService {
	return &SqlStatsService{
		Store:  store,
		Config: conf,
		Logger: logger,
	}
}

func (s *SqlStatsService) QuerySqlStats(req *apimodel.QuerySqlStatsRequest) (*apimodel.SqlStatsResponse, error) {
	filters := s.buildFilters(req)
	s.Logger.Infof("QuerySqlStats filters: %+v", filters)

	selectExpressions, groupByColumns := s.buildQueryParts(req.OutputColumns)

	// Ensure all fixed dimensions are in the SELECT and GROUP BY clauses
	for dim := range fixedDimensions {
		if !contains(groupByColumns, dim) {
			groupByColumns = append(groupByColumns, dim)
		}
		if !contains(selectExpressions, dim) {
			selectExpressions = append(selectExpressions, dim)
		}
	}

	// Create query options for the store
	opts := &store.QueryOptions{
		SelectExpressions: selectExpressions,
		Filters:           filters,
		GroupByColumns:    groupByColumns,
		OrderBy:           req.SortByColumn,
		SortOrder:         req.SortOrder,
		Limit:             req.PageSize,
		Offset:            (req.PageNum - 1) * req.PageSize,
	}

	totalCount, err := s.Store.CountSqlAudits(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to count sql audits: %w", err)
	}
	s.Logger.Infof("QuerySqlStats totalCount: %d", totalCount)

	if totalCount == 0 {
		return &apimodel.SqlStatsResponse{
			Items:      []apimodel.SqlStatsItem{},
			TotalCount: 0,
		}, nil
	}

	results, err := s.Store.QuerySqlAudits(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query sql audits: %w", err)
	}

	items := s.transformResults(results)

	resp := &apimodel.SqlStatsResponse{
		Items:      items,
		TotalCount: totalCount,
	}

	return resp, nil
}

func (s *SqlStatsService) buildQueryParts(outputColumns []string) (selectExpressions []string, groupByColumns []string) {
	for _, col := range outputColumns {
		if expr := common.BuildMetricExpression(col); expr != "" {
			selectExpressions = append(selectExpressions, fmt.Sprintf("%s as %s", expr, col))
		} else if _, isFixedDimension := fixedDimensions[col]; isFixedDimension {
			// It's a dimension
			selectExpressions = append(selectExpressions, col)
			groupByColumns = append(groupByColumns, col)
		} else if _, isDimension := dimensions[col]; isDimension {
			selectExpressions = append(selectExpressions, fmt.Sprintf("MAX(%s) as %s", col, col))
		} else {
			// do nothing for unknown columns
		}
	}
	return
}

func (s *SqlStatsService) transformResults(results []map[string]any) []apimodel.SqlStatsItem {
	items := make([]apimodel.SqlStatsItem, len(results))
	for i, row := range results {
		item := apimodel.SqlStatsItem{
			Statistics: []apimodel.StatisticItem{},
		}
		for key, val := range row {
			// Populate fixed dimensions
			switch key {
			case "svr_ip":
				item.SvrIP, _ = val.(string)
			case "svr_port":
				item.SvrPort, _ = val.(int64)
			case "tenant_id":
				// DuckDB returns BIGINT as int64, need to convert to uint64
				if v, ok := val.(int64); ok {
					item.TenantId = uint64(v)
				}
			case "tenant_name":
				item.TenantName, _ = val.(string)
			case "user_id":
				item.UserId, _ = val.(int64)
			case "user_name":
				item.UserName, _ = val.(string)
			case "db_id":
				if v, ok := val.(int64); ok {
					item.DBId = uint64(v)
				}
			case "db_name":
				item.DBName, _ = val.(string)
			case "sql_id":
				item.SqlId, _ = val.(string)
			case "plan_id":
				item.PlanId, _ = val.(int64)
			case "query_sql":
				item.QuerySql, _ = val.(string)
			case "client_ip":
				item.ClientIp, _ = val.(string)
			case "event":
				item.Event, _ = val.(string)
			case "format_sql_id":
				item.FormatSqlId, _ = val.(string)
			case "effective_tenant_id":
				if v, ok := val.(int64); ok {
					item.EffectiveTenantId = uint64(v)
				}
			case "trace_id":
				item.TraceId, _ = val.(string)
			case "sid":
				if v, ok := val.(int64); ok {
					item.Sid = uint64(v)
				}
			case "user_client_ip":
				item.UserClientIp, _ = val.(string)
			case "tx_id":
				item.TxId, _ = val.(string)
			case "sub_plan_count":
				item.SubPlanCount, _ = val.(int64)
			case "last_fail_info":
				item.LastFailInfo, _ = val.(int64)
			case "cause_type":
				item.CauseType, _ = val.(int64)
			default:
				// If it's a requested metric, add it to the statistics slice
				var floatVal float64
				switch v := val.(type) {
				case int64:
					floatVal = float64(v)
				case float64:
					floatVal = v
				case uint64:
					floatVal = float64(v)
				case *big.Int:
					if v != nil {
						floatVal, _ = v.Float64()
					}
				default:
					// For now, default to 0.0 if type assertion fails
					floatVal = 0.0
				}
				item.Statistics = append(item.Statistics, apimodel.StatisticItem{
					Name:  key,
					Value: floatVal,
				})
			}
		}
		items[i] = item
	}
	return items
}

func (s *SqlStatsService) buildFilters(req *apimodel.QuerySqlStatsRequest) map[string]interface{} {
	filters := make(map[string]interface{})
	if req.StartTime > 0 {
		// The data in parquet is stored as microseconds, so we need to convert
		filters["max_request_time >="] = req.StartTime * 1000000
	}
	if req.EndTime > 0 {
		filters["min_request_time <="] = req.EndTime * 1000000
	}
	if req.UserName != "" {
		filters["user_name ="] = req.UserName
	}
	if req.DatabaseName != "" {
		filters["db_name ="] = req.DatabaseName
	}
	if req.QuerySqlKeyword != "" {
		filters["query_sql ILIKE"] = "%" + req.QuerySqlKeyword + "%"
	}
	if req.FilterInnerSql {
		filters["inner_sql_count ="] = 0
	}
	if req.SuspiciousOnly && s.Config != nil && s.Config.SlowSqlThresholdMilliSeconds > 0 {
		filters["elapsed_time_max >="] = s.Config.SlowSqlThresholdMilliSeconds * 1000
	}
	return filters
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
