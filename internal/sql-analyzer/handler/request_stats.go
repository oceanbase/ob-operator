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

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	logger "github.com/sirupsen/logrus"
)

// @Summary Get request statistics
// @Description Get request statistics
// @Tags sql-analyzer
// @Accept json
// @Produce json
// @Param tenant_name path string true "Tenant Name"
// @Param request body model.RequestStatisticsRequest true "Query parameters"
// @Success 200 {object} model.APIResponse{data=model.RequestStatisticsResponse} "A list of aggregated SQL statistics"
// @Failure 400 {object} model.APIResponse "Error: Invalid request"
// @Failure 500 {object} model.APIResponse "Error: Internal server error"
// @Router /api/v1/tenants/{tenant_name}/request-stats [post]
func GetRequestStatistics(c *gin.Context) (*model.RequestStatisticsResponse, error) {
	var req model.RequestStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "/data"
	}

	l := HandlerLogger
	if l == nil {
		l = logger.StandardLogger()
	}

	auditStore, err := store.NewSqlAuditStore(c.Request.Context(), filepath.Join(dataPath, "sql_audit"), l)
	if err != nil {
		return nil, err
	}
	defer auditStore.Close()

	return business.GetRequestStatistics(auditStore, req)
}
