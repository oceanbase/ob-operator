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

package collector

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
)

// getMaxRequestIDs finds the latest request_id for each observer.
func (c *Collector) getMaxRequestIDs() (map[string]uint64, error) {
	var observers []struct {
		SvrIP        string `db:"svr_ip"`
		MaxRequestID uint64 `db:"max_request_id"`
	}

	cnx, err := c.ConnectionManager.GetSysReadonlyConnection()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get oceanbase connection")
	}

	if err := cnx.QueryList(c.Ctx, &observers, sqlconst.GetMaxRequestIDByIP, c.TenantID, oceanbaseconst.RpcPort); err != nil {
		return nil, errors.Wrap(err, "Failed to query max request ids")
	}

	maxRequestIDs := make(map[string]uint64)
	for _, o := range observers {
		maxRequestIDs[o.SvrIP] = o.MaxRequestID
	}
	return maxRequestIDs, nil
}

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func (c *Collector) collectSqlAuditData() {
	maxRequestIDs, err := c.getMaxRequestIDs()
	if err != nil {
		c.Logger.Errorf("Failed to get max request ids %v", err)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan []model.SqlAudit, len(maxRequestIDs))
	errChan := make(chan error, len(maxRequestIDs))

	for svrIP, maxRequestID := range maxRequestIDs {
		lastRequestID, ok := c.RequestIdMap[svrIP]
		if ok && lastRequestID == maxRequestID {
			continue
		}

		wg.Add(1)
		go func(svrIP string, lastRequestID uint64) {
			defer wg.Done()
			c.Logger.Infof("Collecting from observer %s since request_id %d", svrIP, lastRequestID)
			data, err := c.collectSqlAuditByOBServer(svrIP, lastRequestID)
			if err != nil {
				errChan <- fmt.Errorf("failed to collect from observer %s: %w", svrIP, err)
				return
			}
			if len(data) > 0 {
				resultsChan <- data
			}
		}(svrIP, lastRequestID)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	var allResults [][]model.SqlAudit
	for results := range resultsChan {
		allResults = append(allResults, results)
	}

	for err := range errChan {
		c.Logger.Error("Error during collection:", err) // Log errors but don't fail the whole batch
	}

	totalRecords := 0
	for _, results := range allResults {
		totalRecords += len(results)
		for _, audit := range results {
			c.PushPlan(&model.SqlPlanIdentifier{
				TenantID: c.TenantID,
				SvrIP:    audit.SvrIP,
				SvrPort:  audit.SvrPort,
				PlanID:   audit.PlanId,
			})
			lastRequestID, ok := c.RequestIdMap[audit.SvrIP]
			if !ok || lastRequestID < audit.MaxRequestId {
				c.RequestIdMap[audit.SvrIP] = audit.MaxRequestId
			}
		}
	}
	c.Logger.Infof("Collected %d new audit records.", totalRecords)

	if totalRecords > 0 {
		if err := c.SqlAuditStore.InsertBatch(allResults); err != nil {
			c.Logger.Errorf("Error inserting data into DuckDB: %v", err)
		} else {
			c.Logger.Infof("Saved %d sql audit records", totalRecords)
		}
	}

}

func (c *Collector) PushPlan(plan *model.SqlPlanIdentifier) {
	// Check cache first. This is thread-safe.
	if c.PlanCache.Contains(*plan) {
		c.Logger.Debugf("Plan %v already in cache, skipping.", plan)
		return
	}

	// If not in cache, check DuckDB without holding any lock.
	existsInDuckDB, err := c.SqlPlanStore.PlanExists(*plan)
	if err != nil {
		c.Logger.Errorf("Error checking plan existence in DuckDB for %v: %v", plan, err)
		// If we can't check DuckDB, we'll proceed to collect it, but first add to cache.
	} else if existsInDuckDB {
		c.Logger.Debugf("Plan %v found in DuckDB, adding to cache and skipping.", plan)
		// Add to cache and skip pushing to channel.
		c.PlanCache.Add(*plan, struct{}{})
		return
	}

	// If not in cache and not in DuckDB, then add to cache and push to the channel.
	c.PlanCache.Add(*plan, struct{}{})
	c.PlanIdentifierChan <- plan
}

func (c *Collector) collectSqlAuditByOBServer(svrIP string, lastRequestID uint64) ([]model.SqlAudit, error) {
	var results []model.SqlAudit
	cnx, err := c.ConnectionManager.GetSysReadonlyConnectionByIP(svrIP)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get oceanbase connection")
	}

	if err := cnx.QueryList(c.Ctx, &results, sqlconst.GetSqlStatistics, c.TenantID, svrIP, oceanbaseconst.RpcPort, lastRequestID); err != nil {
		return nil, err
	}
	return results, nil
}
