/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1

import (
	"github.com/gin-gonic/gin"

	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitSqlRoutes(g *gin.RouterGroup) {
	g.GET("/sql/metrics", h.Wrap(h.ListSqlMetrics, acbiz.PathGuard("obcluster", "*", "read")))
	g.POST("/sql/topSqls", h.Wrap(h.ListTopSqls, acbiz.PathGuard("obcluster", "*", "read")))
	g.POST("/sql/suspiciousSqls", h.Wrap(h.ListSuspiciousSqls, acbiz.PathGuard("obcluster", "*", "read")))
	g.POST("/sql/requestStatistics", h.Wrap(h.RequestStatistics, acbiz.PathGuard("obcluster", "*", "read")))
	g.POST("/sql/querySqlDetailInfo", h.Wrap(h.QuerySqlDetailInfo, acbiz.PathGuard("obcluster", "*", "read")))
	g.POST("/sql/queryPlanDetailInfo", h.Wrap(h.QueryPlanDetailInfo, acbiz.PathGuard("obcluster", "*", "read")))
}
