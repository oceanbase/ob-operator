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

	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

// @ID GetSqlDetailInfo
// @Summary Get SQL detail info
// @Description Get SQL detail info
// @Tags SQL
// @Accept application/json
// @Produce application/json
// @Param body body model.SqlDetailRequest true "sql detail request"
// @Success 200 {object} model.SqlDetailResponse
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/v1/stats/sql_detail [POST]
func GetSqlDetailInfo(c *gin.Context) (*model.SqlDetailResponse, error) {
	var req model.SqlDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// Setup logger
	l := HandlerLogger
	if l == nil {
		l = logger.StandardLogger()
	}

	// Setup ConnectionManager
	namespace := os.Getenv("NAMESPACE")
	obTenantName := os.Getenv("OBTENANT")

	var cm *oceanbase.ConnectionManager

	if namespace != "" && obTenantName != "" {
		obtenant, err := clients.GetOBTenant(c.Request.Context(), types.NamespacedName{
			Namespace: namespace,
			Name:      obTenantName,
		})
		if err != nil {
			l.Warnf("Failed to get OBTenant %s/%s: %v", namespace, obTenantName, err)
		} else {
			obcluster, err := clients.GetOBCluster(c.Request.Context(), namespace, obtenant.Spec.ClusterName)
			if err != nil {
				l.Warnf("Failed to get OBCluster %s/%s: %v", namespace, obtenant.Spec.ClusterName, err)
			} else {
				cm = oceanbase.NewConnectionManager(c.Request.Context(), obcluster)
			}
		}
	} else {
		l.Warn("NAMESPACE or OBTENANT env not set, skipping index query")
	}

	return business.GetSqlDetailInfo(c, cm, store.GetSqlAuditStore(), store.GetPlanStore(), req)
}

type DebugQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

func DebugQuery(c *gin.Context) (any, error) {
	var req DebugQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	results, err := store.GetPlanStore().DebugQuery(req.Query)
	if err != nil {
		return nil, err
	}

	return results, nil
}
