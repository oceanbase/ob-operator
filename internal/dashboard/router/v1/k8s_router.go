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

var k8sClusterGuard = acbiz.OR(
	acbiz.PathGuard(string(acbiz.DomainK8sCluster), "*", string(acbiz.ActionRead)),
)

func InitK8sClusterRoutes(g *gin.RouterGroup) {
	g.GET("/k8s/clusters", h.Wrap(h.ListRemoteK8sClusters, k8sClusterGuard))
	g.GET("/k8s/clusters/:name", h.Wrap(h.GetRemoteK8sCluster, k8sClusterGuard))
	g.PUT("/k8s/clusters/:name", h.Wrap(h.GetRemoteK8sCluster, k8sClusterGuard))
	g.POST("/k8s/clusters", h.Wrap(h.CreateRemoteK8sCluster, k8sClusterGuard))
	g.GET("/k8s/clusters/:name/events", h.Wrap(h.ListRemoteK8sEvents, k8sClusterGuard))
	g.GET("/k8s/clusters/:name/nodes", h.Wrap(h.ListRemoteK8sNodes, k8sClusterGuard))
	g.PUT("/k8s/clusters/:clusterName/nodes/:nodeName/labels", h.Wrap(h.PutRemoteK8sNodeLabels, k8sClusterGuard))
	g.PUT("/k8s/clusters/:clusterName/nodes/:nodeName/taints", h.Wrap(h.PutRemoteK8sNodeTaints, k8sClusterGuard))
	g.POST("/k8s/clusters/:name/nodes/update", h.Wrap(h.BatchUpdateRemoteK8sNode, k8sClusterGuard))
}
