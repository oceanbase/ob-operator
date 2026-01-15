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

package handler

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	logger "github.com/sirupsen/logrus"
)

// @Summary Query SQL statistics
// @Description Query SQL statistics data within a given time range and with optional filters.
// @Tags sql-analyzer
// @Accept json
// @Produce json
// @Param tenant_name path string true "Tenant Name"
// @Param request body model.QuerySqlStatsRequest true "Query parameters"
// @Success 200 {object} model.APIResponse{data=model.SqlStatsResponse} "A list of aggregated SQL statistics"
// @Failure 400 {object} model.APIResponse "Error: Invalid request"
// @Failure 500 {object} model.APIResponse "Error: Internal server error"
// @Router /api/v1/tenants/{tenant_name}/sql-stats [post]
func QuerySqlStats(c *gin.Context) (*model.SqlStatsResponse, error) {
	var req model.QuerySqlStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	l := HandlerLogger
	if l == nil {
		l = logger.StandardLogger()
	}
	l.WithField("req", req).Info("QuerySqlStats request")

	// Set default pagination
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "/data"
	}

	auditStore, err := store.NewSqlAuditStore(c.Request.Context(), filepath.Join(dataPath, "sql_audit"), l)
	if err != nil {
		return nil, err
	}
	defer auditStore.Close()

	slowSqlThresholdMilliSeconds := 1000 // milliseconds
	slowSqlThresholdMilliSecondsStr := os.Getenv("SLOW_SQL_THRESHOLD_MILLISECONDS")
	if slowSqlThresholdMilliSecondsStr != "" {
		if val, err := strconv.Atoi(slowSqlThresholdMilliSecondsStr); err == nil && val >= 0 {
			slowSqlThresholdMilliSeconds = val
		}
	}

	conf := &config.Config{
		SlowSqlThresholdMilliSeconds: slowSqlThresholdMilliSeconds,
	}

	service := business.NewSqlStatsService(auditStore, conf, l)
	return service.QuerySqlStats(&req)
}
