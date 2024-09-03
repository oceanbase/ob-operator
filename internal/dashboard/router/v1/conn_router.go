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
	obbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/oceanbase"
	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitTerminalRoutes(g *gin.RouterGroup) {
	g.PUT("/obclusters/namespace/:namespace/name/:name/terminal", h.Wrap(h.CreateOBClusterConnTerminal, acbiz.PathGuard("obcluster", ":namespace+:name", "write")))
	g.PUT("/obtenants/:namespace/:name/terminal", h.Wrap(h.CreateOBTenantConnTerminal, obbiz.TenantGuard(":namespace", ":name", "write")))
	g.GET("/terminal/:terminalId", h.Wrap(h.ConnectDatabase))
}
