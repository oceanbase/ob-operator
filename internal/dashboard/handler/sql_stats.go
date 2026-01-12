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
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/dashboard/client"
	"github.com/oceanbase/ob-operator/internal/dashboard/model"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

func QuerySqlStats(c *gin.Context) (*model.SqlStatsResponse, error) {
	tenantName := c.Param("tenant_name")
	if tenantName == "" {
		return nil, httpErr.NewBadRequest("tenant_name is required")
	}

	var req model.QuerySqlStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, httpErr.NewBadRequest("Invalid request body: " + err.Error())
	}

	sqlAnalyzerClient := client.NewSqlAnalyzerClient()
	resp, err := sqlAnalyzerClient.QuerySqlStats(tenantName, req)
	if err != nil {
		return nil, httpErr.NewInternal("Failed to query sql-stats from sql-analyzer: " + err.Error())
	}

	return resp, nil
}
