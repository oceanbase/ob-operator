/*
Copyright (c) 2023 OceanBase
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

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/clients"
	bizconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/client"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/sql"
	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
)

const (
	SQLMetricConfigFileEnUS = "internal/assets/dashboard/sql_metric_en_US.yaml"
	SQLMetricConfigFileZhCN = "internal/assets/dashboard/sql_metric_zh_CN.yaml"
	SQLMetricScope          = "SQL_DIAGNOSIS"
)

var metricCategoryMap map[string]sql.MetricCategory

func init() {
	metricCategoryMap = make(map[string]sql.MetricCategory)
	metricConfigContent, err := bindata.Asset(SQLMetricConfigFileEnUS)
	if err != nil {
		panic(errors.Wrap(err, "load sql metric config failed"))
	}
	metricConfigs := make([]sql.SqlMetricMetaCategory, 0)
	err = yaml.Unmarshal(metricConfigContent, &metricConfigs)
	if err != nil {
		panic(errors.Wrap(err, "parse sql metric config data failed"))
	}
	for _, category := range metricConfigs {
		for _, metric := range category.Metrics {
			metricCategoryMap[metric.Key] = category.Category
		}
	}
}

func ListSqlMetrics(language string) ([]sql.SqlMetricMetaCategory, error) {
	metricClasses := make([]sql.SqlMetricMetaCategory, 0)
	configFile := SQLMetricConfigFileEnUS
	switch language {
	case bizconstant.LANGUAGE_EN_US:
		configFile = SQLMetricConfigFileEnUS
	case bizconstant.LANGUAGE_ZH_CN:
		configFile = SQLMetricConfigFileZhCN
	default:
		logger.Infof("Not supported language %s, return default", language)
	}

	metricConfigContent, err := bindata.Asset(configFile)
	if err != nil {
		return metricClasses, err
	}
	metricCategories := make([]sql.SqlMetricMetaCategory, 0)
	err = yaml.Unmarshal(metricConfigContent, &metricCategories)
	if err != nil {
		return metricClasses, err
	}
	logger.Debugf("sql metric configs: %v", metricCategories)
	return metricCategories, err
}

func ListSqlStats(ctx context.Context, filter *sql.SqlFilter) (*sql.SqlStatsList, error) {
	sqlAnalyzerAddress, err := k8s.GetSQLAnalyzerAddress(ctx, filter.Namespace, filter.OBTenant)
	if err != nil {
		return nil, err
	}

	req := apimodel.QuerySqlStatsRequest{
		StartTime:       filter.StartTime,
		EndTime:         filter.EndTime,
		UserName:        filter.User,
		DatabaseName:    filter.Database,
		FilterInnerSql:  !filter.IncludeInnerSql,
		SuspiciousOnly:  filter.SuspiciousOnly,
		QuerySqlKeyword: filter.Keyword,
		OutputColumns:   filter.OutputColumns,
		SortByColumn:    filter.SortByColumn,
		SortOrder:       filter.SortOrder,
		PageNum:         filter.PageNum,
		PageSize:        filter.PageSize,
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: filter.Namespace,
		Name:      filter.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	logger.Infof("Sending QuerySqlStatsRequest: %+v", req)
	clt := client.NewClient(fmt.Sprintf("http://%s:8080", sqlAnalyzerAddress))
	resp, err := clt.QuerySqlStats(obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}
	logger.Infof("QuerySqlStats returned %d items", len(resp.Items))

	// Convert resp to []model.SqlInfo
	sqlInfos := make([]sql.SqlInfo, 0, len(resp.Items))
	for _, item := range resp.Items {
		sqlInfo := sql.SqlInfo{
			SqlMetaInfo: sql.SqlMetaInfo{
				SvrIP:      item.SvrIP,
				SvrPort:    item.SvrPort,
				TenantId:   item.TenantId,
				TenantName: item.TenantName,
				UserId:     item.UserId,
				UserName:   item.UserName,
				DBId:       item.DBId,
				DBName:     item.DBName,
				SqlId:      item.SqlId,
				PlanId:     item.PlanId,

				QuerySql:          item.QuerySql,
				ClientIp:          item.ClientIp,
				Event:             item.Event,
				EffectiveTenantId: item.EffectiveTenantId,
				TraceId:           item.TraceId,
				Sid:               item.Sid,
				UserClientIp:      item.UserClientIp,
				TxId:              item.TxId,
				SubPlanCount:      item.SubPlanCount,
				LastFailInfo:      item.LastFailInfo,
				CauseType:         item.CauseType,
			},
			ExecutionStatistics: []sql.SqlStatisticMetric{},
			LatencyStatistics:   []sql.SqlStatisticMetric{},
		}
		for _, stat := range item.Statistics {
			category, ok := metricCategoryMap[stat.Name]
			if !ok {
				logger.Warnf("metric %s has no category", stat.Name)
				continue
			}
			metric := sql.SqlStatisticMetric{
				Name:  stat.Name,
				Value: stat.Value,
			}
			switch category {
			case sql.Execution:
				sqlInfo.ExecutionStatistics = append(sqlInfo.ExecutionStatistics, metric)
			case sql.Latency:
				sqlInfo.LatencyStatistics = append(sqlInfo.LatencyStatistics, metric)
			case sql.Meta:
				// Do nothing, already populated in SqlMetaInfo
			}
		}
		sqlInfos = append(sqlInfos, sqlInfo)
	}

	return &sql.SqlStatsList{
		Items:      sqlInfos,
		TotalCount: resp.TotalCount,
	}, nil
}

func QuerySqlHistoryInfo(ctx context.Context, param *sql.SqlHistoryParam) (*sql.SqlHistoryInfo, error) {
	sqlAnalyzerAddress, err := k8s.GetSQLAnalyzerAddress(ctx, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := apimodel.SqlHistoryRequest{
		StartTime:      param.StartTime,
		EndTime:        param.EndTime,
		SqlId:          param.SqlId,
		Interval:       param.Interval,
		LatencyColumns: param.LatencyColumns,
	}

	clt := client.NewClient(fmt.Sprintf("http://%s:8080", sqlAnalyzerAddress))
	resp, err := clt.QuerySqlHistory(obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil
	}

	historyInfo := &sql.SqlHistoryInfo{
		ExecutionTrend: []response.MetricData{},
		LatencyTrend:   []response.MetricData{},
	}

	// Convert ExecutionTrend
	localTrend := response.MetricData{
		Metric: response.Metric{Name: "local_plan"},
		Values: []response.MetricValue{},
	}
	remoteTrend := response.MetricData{
		Metric: response.Metric{Name: "remote_plan"},
		Values: []response.MetricValue{},
	}
	distributedTrend := response.MetricData{
		Metric: response.Metric{Name: "distributed_plan"},
		Values: []response.MetricValue{},
	}

	for _, trend := range resp.ExecutionTrend {
		ts := float64(trend.Time)
		localTrend.Values = append(localTrend.Values, response.MetricValue{Timestamp: ts, Value: trend.Local})
		remoteTrend.Values = append(remoteTrend.Values, response.MetricValue{Timestamp: ts, Value: trend.Remote})
		distributedTrend.Values = append(distributedTrend.Values, response.MetricValue{Timestamp: ts, Value: trend.Distributed})
	}
	historyInfo.ExecutionTrend = append(historyInfo.ExecutionTrend, localTrend, remoteTrend, distributedTrend)

	// Convert LatencyTrend
	latencyTrends := make(map[string]*response.MetricData)
	for _, col := range param.LatencyColumns {
		latencyTrends[col] = &response.MetricData{
			Metric: response.Metric{Name: col},
			Values: []response.MetricValue{},
		}
	}

	for _, item := range resp.LatencyTrend {
		ts := float64(item.Time)
		for col, val := range item.Value {
			if trend, ok := latencyTrends[col]; ok {
				trend.Values = append(trend.Values, response.MetricValue{Timestamp: ts, Value: val})
			}
		}
	}

	for _, trend := range latencyTrends {
		historyInfo.LatencyTrend = append(historyInfo.LatencyTrend, *trend)
	}

	return historyInfo, nil
}

func QuerySqlDetailInfo(ctx context.Context, param *sql.SqlDetailParam) (*sql.SqlDetailedInfo, error) {
	sqlAnalyzerAddress, err := k8s.GetSQLAnalyzerAddress(ctx, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := apimodel.SqlDetailRequest{
		StartTime: param.StartTime,
		EndTime:   param.EndTime,
		SqlId:     param.SqlId,
	}

	clt := client.NewClient(fmt.Sprintf("http://%s:8080", sqlAnalyzerAddress))
	resp, err := clt.QuerySqlDetail(obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil
	}

	detailedInfo := &sql.SqlDetailedInfo{
		DiagnoseInfo: []sql.SqlDiagnoseInfo{},
		Plans:        []sql.PlanStatistic{},
		Indexies:     []sql.IndexInfo{},
	}

	// Convert Plans
	for _, planStat := range resp.Plans {
		plan := sql.PlanStatistic{
			PlanMeta: sql.PlanMeta{
				PlanIdentity: sql.PlanIdentity{
					TenantID: planStat.TenantID,
					SvrIP:    planStat.SvrIP,
					SvrPort:  planStat.SvrPort,
					PlanID:   planStat.PlanID,
				},
				PlanHash:      planStat.PlanHash,
				GeneratedTime: planStat.GeneratedTime,
			},
			IoCost:   planStat.IoCost,
			CpuCost:  planStat.CpuCost,
			Cost:     planStat.Cost,
			RealCost: planStat.RealCost,
		}
		detailedInfo.Plans = append(detailedInfo.Plans, plan)
	}

	// Convert Indexes
	for _, idx := range resp.Indexes {
		detailedInfo.Indexies = append(detailedInfo.Indexies, sql.IndexInfo{
			TableName:  idx.TableName,
			IndexType:  idx.IndexType,
			Uniqueness: idx.Uniqueness,
			IndexName:  idx.IndexName,
			Columns:    idx.Columns,
			Status:     idx.Status,
		})
	}

	for _, diagnoseInfo := range resp.DiagnoseInfo {
		detailedInfo.DiagnoseInfo = append(detailedInfo.DiagnoseInfo, sql.SqlDiagnoseInfo{
			RuleName:   diagnoseInfo.RuleName,
			Level:      diagnoseInfo.Level,
			Reason:     diagnoseInfo.Reason,
			Suggestion: diagnoseInfo.Suggestion,
		})
	}

	return detailedInfo, nil
}

func ListRequestStatistics(c context.Context, param *sql.SqlRequestStatisticParam) ([]sql.RequestStatisticInfo, error) {
	sqlAnalyzerAddress, err := k8s.GetSQLAnalyzerAddress(c, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(c, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := apimodel.RequestStatisticsRequest{
		StartTime:      param.StartTime,
		EndTime:        param.EndTime,
		UserName:       param.User,
		DatabaseName:   param.Database,
		FilterInnerSql: !param.IncludeInnerSql,
	}

	clt := client.NewClient(fmt.Sprintf("http://%s:8080", sqlAnalyzerAddress))
	resp, err := clt.QueryRequestStatistics(obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return []sql.RequestStatisticInfo{}, nil
	}

	var averageLatency float64
	if resp.TotalExecutions > 0 {
		averageLatency = resp.TotalLatency / resp.TotalExecutions
	}

	info := sql.RequestStatisticInfo{
		Tenant:                 obtenant.Spec.TenantName,
		User:                   param.User,
		Database:               param.Database,
		PlanCategoryStatistics: []sql.SqlStatisticMetric{}, // This field is not available from the sql-analyzer
		TotalExecutions:        resp.TotalExecutions,
		FailedExecutions:       resp.FailedExecutions,
		TotalLatency:           resp.TotalLatency,
		AverageLatency:         averageLatency,
		ExecutionTrend:         []response.MetricValue{},
		LatencyTrend:           []response.MetricValue{},
	}

	for _, trend := range resp.ExecutionTrend {
		t, err := time.Parse("2006-01-02", trend.Date)
		if err != nil {
			logger.Errorf("Failed to parse date string %s: %v", trend.Date, err)
			continue
		}
		timestamp := float64(t.Unix())
		info.ExecutionTrend = append(info.ExecutionTrend, response.MetricValue{
			Timestamp: timestamp,
			Value:     trend.Value,
		})
	}

	for _, trend := range resp.LatencyTrend {
		t, err := time.Parse("2006-01-02", trend.Date)
		if err != nil {
			logger.Errorf("Failed to parse date string %s: %v", trend.Date, err)
			continue
		}
		timestamp := float64(t.Unix())
		info.LatencyTrend = append(info.LatencyTrend, response.MetricValue{
			Timestamp: timestamp,
			Value:     trend.Value,
		})
	}

	return []sql.RequestStatisticInfo{info}, nil
}

func QueryPlanDetailInfo(ctx context.Context, param *sql.PlanDetailParam) (*sql.PlanDetail, error) {
	sqlAnalyzerAddress, err := k8s.GetSQLAnalyzerAddress(ctx, param.Namespace, param.OBTenant)
	if err != nil {
		return nil, err
	}

	obtenant, err := clients.GetOBTenant(ctx, types.NamespacedName{
		Namespace: param.Namespace,
		Name:      param.OBTenant,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get ob tenant")
	}

	req := model.SqlPlanIdentifier{
		TenantID: param.TenantID,
		SvrIP:    param.SvrIP,
		SvrPort:  param.SvrPort,
		PlanID:   param.PlanID,
	}

	clt := client.NewClient(fmt.Sprintf("http://%s:8080", sqlAnalyzerAddress))
	plans, err := clt.QueryPlanDetail(obtenant.Spec.TenantName, req)
	if err != nil {
		return nil, err
	}

	if len(plans) == 0 {
		return nil, nil
	}

	// Build plan tree
	planMap := make(map[int64]*sql.PlanOperator)
	var root *sql.PlanOperator

	for _, plan := range plans {
		outputOrFilter := make([]string, 0)
		if plan.AccessPredicates != "" {
			outputOrFilter = append(outputOrFilter, fmt.Sprintf("access: %s", plan.AccessPredicates))
		}
		if plan.FilterPredicates != "" {
			outputOrFilter = append(outputOrFilter, fmt.Sprintf("filter: %s", plan.FilterPredicates))
		}
		if plan.StartupPredicates != "" {
			outputOrFilter = append(outputOrFilter, fmt.Sprintf("startup: %s", plan.StartupPredicates))
		}
		if plan.Projection != "" {
			outputOrFilter = append(outputOrFilter, fmt.Sprintf("projection: %s", plan.Projection))
		}
		if plan.SpecialPredicates != "" {
			outputOrFilter = append(outputOrFilter, fmt.Sprintf("special: %s", plan.SpecialPredicates))
		}
		planMap[plan.ID] = &sql.PlanOperator{
			Operator:       plan.Operator,
			Name:           plan.ObjectName,
			EstimatedRows:  int(plan.Cardinality),
			Cost:           plan.Cost,
			OutputOrFilter: strings.Join(outputOrFilter, "\n"),
		}
	}

	for _, plan := range plans {
		if plan.ParentID == -1 {
			root = planMap[plan.ID]
		} else {
			parent, ok := planMap[plan.ParentID]
			if ok {
				parent.ChildOperators = append(parent.ChildOperators, planMap[plan.ID])
			}
		}
	}

	planIdentity := sql.PlanIdentity{
		SvrIP:    plans[0].SvrIP,
		SvrPort:  plans[0].SvrPort,
		TenantID: plans[0].TenantID,
		PlanID:   plans[0].PlanID,
	}

	return &sql.PlanDetail{
		PlanMeta: sql.PlanMeta{
			PlanIdentity: planIdentity,
			PlanHash:     plans[0].PlanHash,
		},
		PlanDetail: root,
	}, nil
}
