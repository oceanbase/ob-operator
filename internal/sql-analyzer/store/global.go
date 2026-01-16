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

// Package store provides data storage and retrieval for SQL analysis results.
package store

import (
	"context"
	"path/filepath"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/sirupsen/logrus"
)

var (
	globalSqlAuditStore *SqlAuditStore
	globalPlanStore     *PlanStore
)

func InitGlobalStores(ctx context.Context, conf *config.Config, logger *logrus.Logger) error {
	var err error
	globalSqlAuditStore, err = NewSqlAuditStore(ctx, filepath.Join(conf.DataPath, "sql_audit"), conf.DuckDBMaxOpenConns, logger)
	if err != nil {
		return err
	}

	globalPlanStore, err = NewPlanStore(ctx, filepath.Join(conf.DataPath, "sql_plan"), conf.DuckDBMaxOpenConns, logger)
	if err != nil {
		return err
	}
	return nil
}

func GetSqlAuditStore() *SqlAuditStore {
	return globalSqlAuditStore
}

func GetPlanStore() *PlanStore {
	return globalPlanStore
}

func CloseGlobalStores() {
	if globalSqlAuditStore != nil {
		globalSqlAuditStore.Close()
	}
	if globalPlanStore != nil {
		globalPlanStore.Close()
	}
}
