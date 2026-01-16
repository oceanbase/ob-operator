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
	"fmt" // Added import
	"path/filepath"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

type Collector struct {
	Ctx                context.Context
	Config             *config.Config
	ConnectionManager  *oceanbase.ConnectionManager
	SqlAuditStore      *store.SqlAuditStore
	SqlPlanStore       *store.PlanStore
	RequestIdMap       map[string]uint64
	PlanCache          *lru.Cache[model.SqlPlanIdentifier, struct{}]
	TenantID           uint64
	PlanIdentifierChan chan *model.SqlPlanIdentifier
	CompactionChan     chan struct{}
	Logger             *logrus.Logger
}

// NewCollector creates a new Collector.
func NewCollector(ctx context.Context, config *config.Config, logger *logrus.Logger) *Collector {
	c := &Collector{
		Ctx:                ctx,
		Config:             config,
		PlanIdentifierChan: make(chan *model.SqlPlanIdentifier, config.QueueSize),
		CompactionChan:     make(chan struct{}, 1),
		Logger:             logger,
	}
	var err error
	c.PlanCache, err = lru.New[model.SqlPlanIdentifier, struct{}](config.PlanCacheSize)
	if err != nil {
		logger.Fatalf("Failed to create LRU cache: %v", err)
	}
	return c
}

func (c *Collector) Init() error {
	sqlAuditStore, err := store.NewSqlAuditStore(c.Ctx, filepath.Join(c.Config.DataPath, "sql_audit"), c.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize sql audit store: %w", err)
	}
	c.SqlAuditStore = sqlAuditStore

	planStore, err := store.NewPlanStore(c.Ctx, filepath.Join(c.Config.DataPath, "sql_plan"), c.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize sql plan store: %w", err)
	}
	c.SqlPlanStore = planStore

	err = c.SqlPlanStore.InitSqlPlanTable()
	if err != nil {
		c.SqlPlanStore.Close()
		return errors.Wrap(err, "Failed to init sql plan table")
	}

	c.SqlAuditStore.StartCleanupWorker()

	obtenant, err := clients.GetOBTenant(c.Ctx, types.NamespacedName{
		Namespace: c.Config.Namespace,
		Name:      c.Config.OBTenant,
	})

	if err != nil {
		return fmt.Errorf("failed to get OBTenant resource: %w", err)
	}

	// Get the OBCluster resource.
	obcluster, err := clients.GetOBCluster(c.Ctx, c.Config.Namespace, obtenant.Spec.ClusterName)
	if err != nil {
		return fmt.Errorf("failed to get OBCluster resource: %w", err)
	}

	connectionManager := oceanbase.NewConnectionManager(c.Ctx, obcluster)
	c.ConnectionManager = connectionManager

	lastRequestIDs, err := sqlAuditStore.GetLastRequestIDs()
	if err != nil {
		return fmt.Errorf("failed to load request id from duckdb: %w", err)
	} else {
		for k, v := range lastRequestIDs {
			c.Logger.Infof("Retrieved progress for %s with request id %d from DuckDB.", k, v)
		}
	}
	c.RequestIdMap = lastRequestIDs

	// what if there's a huge number of plans
	existingPlans, err := planStore.LoadExistingPlans()
	if err != nil {
		return fmt.Errorf("failed to load plan identities from duckdb: %w", err)
	}

	for _, plan := range existingPlans {
		c.PlanCache.Add(plan, struct{}{}) // Add with struct{} as value
	}

	tenantID, err := getTenantIDByName(c.Ctx, connectionManager, obtenant.Spec.TenantName)
	if err != nil {
		return fmt.Errorf("failed to get tenant id from oceanbase: %w", err)
	}
	c.TenantID = tenantID
	return nil
}

func (c *Collector) Stop() {
	defer c.SqlAuditStore.Close()
	defer c.SqlPlanStore.Close()
}

func (c *Collector) Start() {
	var wg sync.WaitGroup
	wg.Add(c.Config.WorkerNum)

	for i := 0; i < c.Config.WorkerNum; i++ {
		worker := NewPlanWorker(c, c.ConnectionManager, c.PlanIdentifierChan, c.SqlPlanStore, &wg) // Pass 'c'
		go worker.Start(c.Ctx, i)
	}

	// Start compaction worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-c.CompactionChan:
				c.Logger.Info("Compaction signal received, running compaction...")
				if err := c.SqlAuditStore.Compact(); err != nil {
					c.Logger.Errorf("Failed to compact sql audit data: %v", err)
				} else {
					c.Logger.Info("Sql audit data compacted successfully.")
				}
			case <-c.Ctx.Done():
				c.Logger.Info("Compaction worker stopped.")
				return
			}
		}
	}()

	// Run the collection loop.
	ticker := time.NewTicker(c.Config.Interval)
	defer ticker.Stop()

	// Run a collection immediately at startup.
	compactionCounter := 0

	for {
		select {
		case <-ticker.C:
			c.collectSqlAuditData()
			compactionCounter++
			if compactionCounter >= c.Config.CompactionThreshold {
				select {
				case c.CompactionChan <- struct{}{}: // Send compaction signal
					compactionCounter = 0 // Reset counter after sending signal
				default:
					c.Logger.Warn("Compaction channel is full, skipping compaction signal.")
				}
			}
		case <-c.Ctx.Done():
			c.Logger.Info("Collector stopped. Stopping plan workers...")
			close(c.PlanIdentifierChan)
			wg.Wait() // Wait for all workers (plan and compaction) to finish
			return
		}
	}
}
