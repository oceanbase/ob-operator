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
	"context"
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/analyzer"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetSqlDetailInfo(ctx context.Context, cm *oceanbase.ConnectionManager, auditStore *store.SqlAuditStore, planStore *store.PlanStore, req model.SqlDetailRequest) (*model.SqlDetailResponse, error) {
	start := time.Now()
	resp, err := auditStore.QuerySqlDetailInfo(planStore, req)
	if err != nil {
		return nil, err
	}
	logger.Infof("[GetSqlDetailInfo] QuerySqlDetailInfo took %v", time.Since(start))

	// If we have tables and connection manager, query indexes
	if resp != nil && len(resp.Tables) > 0 && cm != nil {
		indexStart := time.Now()
		opMgr, err := cm.GetSysReadonlyConnection()
		if err != nil {
			logger.Warnf("Failed to get sys connection for index query: %v", err)
			// Do not fail immediately, continue to analysis with empty indexes if needed or proceed
		} else {
			defer opMgr.Close()

			// We need tenantID of the USER tenant, not sys tenant.
			// resp.Plans has TenantID. It should be the same for all plans of the same SQL usually, or at least we pick one.
			var tenantID uint64
			if len(resp.Plans) > 0 {
				tenantID = resp.Plans[0].TenantID
			} else {
				logger.Warn("No plans found, cannot determine tenantID for index query")
			}

			if tenantID > 0 {
				var mu sync.Mutex
				var wg sync.WaitGroup

				for _, table := range resp.Tables {
					wg.Add(1)
					go func(t model.TableInfo) {
						defer wg.Done()
						tableStart := time.Now()
						indexes, err := oceanbase.QueryTableIndexes(ctx, opMgr, tenantID, t.DatabaseName, t.TableName, t.TableID)
						if err != nil {
							logger.Warnf("Failed to query indexes for table %s.%s: %v", t.DatabaseName, t.TableName, err)
							return
						}

						mu.Lock()
						resp.Indexes = append(resp.Indexes, indexes...)
						mu.Unlock()

						logger.Infof("[GetSqlDetailInfo] QueryTableIndexes for %s.%s with table id %d took %v", t.DatabaseName, t.TableName, t.TableID, time.Since(tableStart))
					}(table)
				}
				wg.Wait()
			}
		}
		logger.Infof("[GetSqlDetailInfo] Query Indexes total took %v", time.Since(indexStart))
	}

	// Initialize the SQL Analyzer and run analysis
	// Analyze now requires Indexes
	analyzerManager := analyzer.NewManager()
	if resp != nil && resp.QuerySql != "" {
		analyzeStart := time.Now()
		diagnoseResults := analyzerManager.Analyze(resp.QuerySql, resp.Indexes)
		resp.DiagnoseInfo = diagnoseResults
		logger.Infof("[GetSqlDetailInfo] Analyze took %v", time.Since(analyzeStart))
	} else {
		if resp == nil {
			logger.Warn("[GetSqlDetailInfo] Response is nil, skipping analysis")
		} else {
			logger.Warnf("[GetSqlDetailInfo] QuerySql is empty for sqlId %s, skipping analysis. StartTime: %d, EndTime: %d", req.SqlId, req.StartTime, req.EndTime)
		}
	}

	logger.Infof("[GetSqlDetailInfo] Total execution time: %v", time.Since(start))
	return resp, nil
}
