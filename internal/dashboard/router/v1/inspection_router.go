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

func InitInspectionRoutes(g *gin.RouterGroup) {
	g.GET("/inspection/policies", h.Wrap(h.ListInspectionPolicies, acbiz.PathGuard("obcluster", "*", "read")))
	g.GET("/inspection/policies/:namespace/:name", h.Wrap(h.GetInspectionPolicy, acbiz.PathGuard("obcluster", "*", "read")))
	g.POST("/inspection/policies", h.Wrap(h.CreateOrUpdateInspectionPolicy, acbiz.PathGuard("obcluster", "*", "write")))
	g.DELETE("/inspection/policies/:namespace/:name/:scenario", h.Wrap(h.DeleteInspectionPolicy, acbiz.PathGuard("obcluster", "*", "write")))
	g.POST("/inspection/policies/:namespace/:name/:scenario/trigger", h.Wrap(h.TriggerInspection, acbiz.PathGuard("obcluster", "*", "write")))
	g.GET("/inspection/reports", h.Wrap(h.ListInspectionReports, acbiz.PathGuard("obcluster", "*", "read")))
	g.GET("/inspection/reports/:namespace/:name", h.Wrap(h.GetInspectionReport, acbiz.PathGuard("obcluster", "*", "read")))
}
