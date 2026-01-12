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
	"context"
	"sync"

	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

type PlanWorker struct {
	collector   *Collector // New field
	connManager *oceanbase.ConnectionManager
	inputChan   chan *model.SqlPlanIdentifier
	planStore   *store.PlanStore
	wg          *sync.WaitGroup
}

// NewPlanWorker creates a new PlanWorker.
func NewPlanWorker(collector *Collector, connManager *oceanbase.ConnectionManager, inputChan chan *model.SqlPlanIdentifier, planStore *store.PlanStore, wg *sync.WaitGroup) *PlanWorker {
	return &PlanWorker{
		collector:   collector, // Assign the collector
		connManager: connManager,
		inputChan:   inputChan,
		planStore:   planStore,
		wg:          wg,
	}
}

// Start starts the worker.
func (w *PlanWorker) Start(ctx context.Context, idx int) {
	w.collector.Logger.Infof("plan worker %d started", idx)
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ident := <-w.inputChan:
			w.processPlan(ctx, idx, ident)
		}
	}
}

func (w *PlanWorker) processPlan(ctx context.Context, idx int, ident *model.SqlPlanIdentifier) {
	w.collector.Logger.Printf("Fetching plan for tenant %d, server %s, port %d, plan %d in worker %d", ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID, idx)
	cnx, err := w.connManager.GetSysReadonlyConnectionByIP(ident.SvrIP)
	if err != nil {
		w.collector.Logger.Printf("failed to get connection for plan worker: %v", err)
		// Remove from cache if failed
		w.collector.PlanCache.Remove(*ident) // Remove from cache
		return
	}
	defer cnx.Close()

	var plans []model.SqlPlan
	if err := cnx.QueryList(ctx, &plans, sqlconst.SelectSqlPlan, ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID); err != nil {
		w.collector.Logger.Printf("failed to query sql plan: %v", err)
		// Remove from cache if failed
		w.collector.PlanCache.Remove(*ident) // Remove from cache
		return
	}
	w.collector.Logger.Printf("Found %d plan details for tenant %d, server %s, port %d, plan %d", len(plans), ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID)
	allStored := true
	for _, plan := range plans {
		if err := w.planStore.Store(plan); err != nil {
			w.collector.Logger.WithField("plan", plan).Errorf("Error inserting plan into DuckDB: %v", err)
			allStored = false
			break // Stop processing further plans for this identifier if one fails
		}
	}
	// Update cache status based on storage result
	if allStored {
		w.collector.PlanCache.Add(*ident, struct{}{}) // Add to cache with empty struct
	} else {
		// If not all stored, remove from cache
		w.collector.PlanCache.Remove(*ident) // Remove from cache
	}
}
