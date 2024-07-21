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

	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitAccessControlRoutes(g *gin.RouterGroup) {
	g.GET("/ac/accounts", h.Wrap(h.ListAccounts))
	g.POST("/ac/accounts", h.Wrap(h.CreateAccount))
	g.PATCH("/ac/accounts/:username", h.Wrap(h.PatchAccount))
	g.DELETE("/ac/accounts/:username", h.Wrap(h.DeleteAccount))

	g.GET("/ac/roles", h.Wrap(h.ListRoles))
	g.POST("/ac/roles", h.Wrap(h.CreateRole))
	g.PATCH("/ac/roles/:name", h.Wrap(h.PatchRole))
	g.DELETE("/ac/roles/:name", h.Wrap(h.DeleteRole))

	g.GET("/ac/info", h.Wrap(h.GetAccountInfo))
	g.GET("/ac/policies", h.Wrap(h.ListAllPolicies))
}
