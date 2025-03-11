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

// Users with read permission for cluster and obproxy can list events, too
var eventGuard = acbiz.OR(
	acbiz.PathGuard(string(acbiz.DomainSystem), "*", string(acbiz.ActionRead)),
	acbiz.PathGuard(string(acbiz.DomainOBCluster), "*", string(acbiz.ActionRead)),
	acbiz.PathGuard(string(acbiz.DomainOBProxy), "*", string(acbiz.ActionRead)),
)

// Users with write permission for cluster and obproxy can list namespaces and storage classes
var k8sResourceGuard = acbiz.OR(
	acbiz.PathGuard(string(acbiz.DomainSystem), "*", string(acbiz.ActionRead)),
	acbiz.PathGuard(string(acbiz.DomainOBCluster), "*", string(acbiz.ActionWrite)),
	acbiz.PathGuard(string(acbiz.DomainOBProxy), "*", string(acbiz.ActionWrite)),
)

func InitK8sRoutes(g *gin.RouterGroup) {
	g.GET("/cluster/events", h.Wrap(h.ListK8sEvents, eventGuard))
	g.GET("/cluster/nodes", h.Wrap(h.ListK8sNodes, acbiz.PathGuard("system", "*", "read")))
	g.PUT("/cluster/nodes/:name/labels", h.Wrap(h.PutK8sNodeLabels, acbiz.PathGuard("system", "*", "write")))
	g.PUT("/cluster/nodes/:name/taints", h.Wrap(h.PutK8sNodeTaints, acbiz.PathGuard("system", "*", "write")))
	g.POST("/cluster/nodes/update", h.Wrap(h.BatchUpdateK8sNodes, acbiz.PathGuard("system", "*", "write")))
	g.GET("/cluster/namespaces", h.Wrap(h.ListK8sNamespaces, k8sResourceGuard))
	g.GET("/cluster/storageClasses", h.Wrap(h.ListK8sStorageClasses, k8sResourceGuard))
	g.POST("/cluster/namespaces", h.Wrap(h.CreateK8sNamespace, acbiz.PathGuard("system", "*", "write")))
}
