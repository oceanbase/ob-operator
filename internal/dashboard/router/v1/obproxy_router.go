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

func InitOBProxyRoutes(g *gin.RouterGroup) {
	g.GET("/obproxies", h.Wrap(h.ListOBProxies, acbiz.PathGuard("obproxy", "*", "read")))
	g.PUT("/obproxies", h.Wrap(h.CreateOBProxy, acbiz.PathGuard("obproxy", "*", "write")))
	g.GET("/obproxies/:namespace/:name", h.Wrap(h.GetOBProxy, acbiz.PathGuard("obproxy", ":namespace+:name", "read")))
	g.PATCH("/obproxies/:namespace/:name", h.Wrap(h.PatchOBProxy, acbiz.PathGuard("obproxy", ":namespace+:name", "write")))
	g.DELETE("/obproxies/:namespace/:name", h.Wrap(h.DeleteOBProxy, acbiz.PathGuard("obproxy", ":namespace+:name", "write")))
	g.GET("/obproxies/:namespace/:name/parameters", h.Wrap(h.ListOBProxyParameters, acbiz.PathGuard("obproxy", ":namespace+:name", "read")))
}
