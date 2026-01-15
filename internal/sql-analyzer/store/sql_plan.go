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

package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	logger "github.com/sirupsen/logrus"
)

type PlanStore struct {
	ctx    context.Context
	db     *sql.DB
	mu     sync.RWMutex
	Logger *logger.Logger
}

func (s *PlanStore) InitSqlPlanTable() error {
	// Create table if not exists
	_, err := s.db.Exec(sqlconst.CreateSqlPlanTable)
	return err
}

func NewPlanStore(c context.Context, path string, readOnly bool, l *logger.Logger) (*PlanStore, error) {
	l.Infof("Using plan store at %s", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", path, err)
	}

	dsn := filepath.Join(path, "sql_plan.duckdb")
	var db *sql.DB
	var err error
	var conn *sql.Conn

	// Retry loop to handle database lock during rolling updates
	for i := 0; i < 30; i++ {
		db, err = sql.Open("duckdb", dsn)
		if err != nil {
			l.Warnf("Failed to acquire lock on duckdb, retrying in 2 seconds... Error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// sql.Open doesn't actually connect. We need to try to get a connection.
		conn, err = db.Conn(c)
		if err == nil {
			break // Success
		}

		db.Close() // Close the db handle on failure
		l.Warnf("Failed to acquire lock on duckdb, retrying in 2 seconds... Error: %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to open duckdb after retries at path %s", path)
	}

	// Set memory limit
	memLimit := os.Getenv("DUCKDB_MEMORY_LIMIT")
	if memLimit == "" {
		memLimit = "512MB"
	}
	if _, err := conn.ExecContext(c, fmt.Sprintf("PRAGMA memory_limit='%s'", memLimit)); err != nil {
		l.Warnf("Failed to set duckdb memory limit: %v", err)
	}

	conn.Close() // Close the temporary connection, the pool will manage connections from here.

	s := &PlanStore{db: db, ctx: c, Logger: l}
	return s, nil
}

func (s *PlanStore) LoadExistingPlans() ([]model.SqlPlanIdentifier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(sqlconst.ListSqlPlanIdentifier)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query existing plans")
	}
	defer rows.Close()

	existingPlans := make([]model.SqlPlanIdentifier, 0)
	for rows.Next() {
		var tenantID uint64
		var svrIP string
		var svrPort int64
		var planID int64
		if err := rows.Scan(&tenantID, &svrIP, &svrPort, &planID); err != nil {
			return nil, errors.Wrap(err, "failed to scan existing plan")
		}
		existingPlans = append(existingPlans, model.SqlPlanIdentifier{
			TenantID: tenantID,
			SvrIP:    svrIP,
			SvrPort:  svrPort,
			PlanID:   planID,
		})
	}
	return existingPlans, nil
}

func (s *PlanStore) Store(plan model.SqlPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	valueArgs := []interface{}{plan.TenantID, plan.SvrIP, plan.SvrPort, plan.PlanID, plan.SqlID, plan.DbID, fmt.Sprintf("%d", plan.PlanHash), plan.GmtCreate,
		plan.Operator, plan.ObjectNode, plan.ObjectID, plan.ObjectOwner, plan.ObjectName, plan.ObjectAlias,
		plan.ObjectType, plan.Optimizer, plan.ID, plan.ParentID, plan.Depth, plan.Position, plan.Cost, plan.RealCost,
		plan.Cardinality, plan.RealCardinality, plan.IoCost, plan.CpuCost, plan.Bytes, plan.Rowset, plan.OtherTag,
		plan.PartitionStart, plan.Other, plan.AccessPredicates, plan.FilterPredicates, plan.StartupPredicates,
		plan.Projection, plan.SpecialPredicates, plan.QblockName, plan.Remarks, plan.OtherXML}

	if _, err := s.db.Exec(sqlconst.StoreSqlPlanStatement, valueArgs...); err != nil {
		return err
	}

	return nil
}

func (s *PlanStore) PlanExists(ident model.SqlPlanIdentifier) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	err := s.db.QueryRow(sqlconst.CheckPlanExistence, ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "failed to query plan existence")
	}
	return count > 0, nil
}

func (s *PlanStore) GetPlanDetail(ident model.SqlPlanIdentifier) ([]model.SqlPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(sqlconst.SelectSqlPlanFromDuckdb, ident.TenantID, ident.SvrIP, ident.SvrPort, ident.PlanID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plans by sqlId and planHash")
	}
	defer rows.Close()

	plans := make([]model.SqlPlan, 0)
	for rows.Next() {
		var plan model.SqlPlan
		var planHashStr string
		if err := rows.Scan(&plan.TenantID, &plan.SvrIP, &plan.SvrPort, &plan.PlanID, &plan.SqlID, &plan.DbID, &planHashStr, &plan.GmtCreate,
			&plan.Operator, &plan.ObjectNode, &plan.ObjectID, &plan.ObjectOwner, &plan.ObjectName, &plan.ObjectAlias,
			&plan.ObjectType, &plan.Optimizer, &plan.ID, &plan.ParentID, &plan.Depth, &plan.Position, &plan.Cost, &plan.RealCost,
			&plan.Cardinality, &plan.RealCardinality, &plan.IoCost, &plan.CpuCost, &plan.Bytes, &plan.Rowset, &plan.OtherTag,
			&plan.PartitionStart, &plan.Other, &plan.AccessPredicates, &plan.FilterPredicates, &plan.StartupPredicates,
			&plan.Projection, &plan.SpecialPredicates, &plan.QblockName, &plan.Remarks, &plan.OtherXML); err != nil {
			return nil, errors.Wrap(err, "failed to scan plan")
		}
		plan.PlanHash, err = strconv.ParseUint(planHashStr, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse plan hash")
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// Close closes the database connection.
func (s *PlanStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *PlanStore) GetPlanStatsBySqlId(sqlId string) ([]model.PlanStatistic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(sqlconst.GetPlanStats, sqlId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query plan statistics by sqlId")
	}
	defer rows.Close()

	stats := make([]model.PlanStatistic, 0)
	for rows.Next() {
		var stat model.PlanStatistic
		var planHashStr string
		if err := rows.Scan(
			&stat.TenantID,
			&stat.SvrIP,
			&stat.SvrPort,
			&stat.PlanID,
			&planHashStr,
			&stat.GeneratedTime,
			&stat.IoCost,
			&stat.CpuCost,
			&stat.Cost,
			&stat.RealCost,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan plan statistic")
		}
		stat.PlanHash, err = strconv.ParseUint(planHashStr, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse plan hash in stats")
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

func (s *PlanStore) GetTableInfoBySqlId(sqlId string) ([]model.TableInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// keep the max object id(table id), tables may dropped and recreated
	rows, err := s.db.Query(sqlconst.GetTableInfo, sqlId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query table info by sqlId")
	}
	defer rows.Close()

	tables := make([]model.TableInfo, 0)
	for rows.Next() {
		var table model.TableInfo
		if err := rows.Scan(
			&table.DatabaseName,
			&table.TableName,
			&table.TableID,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan table info")
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (s *PlanStore) DebugQuery(query string, args ...interface{}) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if val != nil {
				if b, ok := (*val).([]byte); ok {
					m[colName] = string(b)
				} else {
					m[colName] = *val
				}
			} else {
				m[colName] = nil
			}
		}
		results = append(results, m)
	}
	return results, nil
}
