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

	h "github.com/oceanbase/ob-operator/internal/dashboard/handler"
)

func InitK8sRoutes(g *gin.RouterGroup) {
	g.GET("/cluster/events", h.Wrap(h.ListK8sEvents))
	g.GET("/cluster/nodes", h.Wrap(h.ListK8sNodes))
	g.GET("/cluster/namespaces", h.Wrap(h.ListK8sNamespaces))
	g.GET("/cluster/storageClasses", h.Wrap(h.ListK8sStorageClasses))
	g.POST("/cluster/namespaces", h.Wrap(h.CreateK8sNamespace))
}
