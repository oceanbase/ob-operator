/*
Copyright (c) 2024 OceanBase
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

func InitAccessControlRoutes(g *gin.RouterGroup) {
	g.GET("/ac/accounts", h.Wrap(h.ListAccounts, acbiz.PathGuard("ac", "*", "read")))
	g.POST("/ac/accounts", h.Wrap(h.CreateAccount, acbiz.PathGuard("ac", "*", "write")))
	g.PATCH("/ac/accounts/:username", h.Wrap(h.PatchAccount, acbiz.PathGuard("ac", "*", "write")))
	g.DELETE("/ac/accounts/:username", h.Wrap(h.DeleteAccount, acbiz.PathGuard("ac", "*", "write")))

	g.GET("/ac/roles", h.Wrap(h.ListRoles, acbiz.PathGuard("ac", "*", "read")))
	g.POST("/ac/roles", h.Wrap(h.CreateRole, acbiz.PathGuard("ac", "*", "write")))
	g.PATCH("/ac/roles/:name", h.Wrap(h.PatchRole, acbiz.PathGuard("ac", "*", "write")))
	g.DELETE("/ac/roles/:name", h.Wrap(h.DeleteRole, acbiz.PathGuard("ac", "*", "write")))

	g.GET("/ac/info", h.Wrap(h.GetAccountInfo))
	g.GET("/ac/policies", h.Wrap(h.ListAllPolicies, acbiz.PathGuard("ac", "*", "read")))
}
