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

package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/obproxy"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/k8s/resource"
	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ListK8sClusterEvents(ctx context.Context, c *client.Client, queryEventParam *param.QueryEventParam) ([]response.K8sEvent, error) {
	events := make([]response.K8sEvent, 0)
	listOptions := &metav1.ListOptions{}
	var selectors []string
	if queryEventParam.Name != "" {
		selectors = append(selectors, fmt.Sprintf("involvedObject.name=%s", queryEventParam.Name))
	}
	if queryEventParam.ObjectType != "" {
		var kind string
		switch queryEventParam.ObjectType {
		case "OBCLUSTER":
			kind = "OBCluster"
		case "OBTENANT":
			kind = "OBTenant"
		case "OBBACKUPPOLICY":
			kind = "OBTenantBackupPolicy"
		case "OBPROXY":
			kind = "Deployment"
		default:
			kind = queryEventParam.ObjectType
		}
		selectors = append(selectors, fmt.Sprintf("involvedObject.kind=%s", kind))
	}
	if queryEventParam.Type != "" {
		var eventType string
		switch queryEventParam.Type {
		case "NORMAL":
			eventType = "Normal"
		case "WARNING":
			eventType = "Warning"
		}
		selectors = append(selectors, fmt.Sprintf("type=%s", eventType))
	}
	ns := corev1.NamespaceAll
	if queryEventParam.Namespace != "" {
		ns = queryEventParam.Namespace
	}
	if len(selectors) > 0 {
		listOptions.FieldSelector = strings.Join(selectors, ",")
	}
	var filterMap map[string]struct{}
	if queryEventParam.ObjectType == "OBPROXY" {
		// Filter events by obproxy deployments
		filterMap = make(map[string]struct{})
		deployments, err := c.MetaClient.Resource(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}).Namespace(ns).List(ctx, metav1.ListOptions{
			LabelSelector: obproxy.LabelOBProxy,
		})
		logger.Debugf("List deployments: %+v", deployments.Items)
		if err != nil {
			return nil, err
		}
		for _, deploy := range deployments.Items {
			filterMap[deploy.Name] = struct{}{}
		}
	}
	eventList, err := resource.ListEvents(ctx, ns, listOptions)
	logger.Infof("Query events with param: %+v", queryEventParam)
	if err == nil {
		for _, event := range eventList.Items {
			if filterMap != nil {
				if _, ok := filterMap[event.InvolvedObject.Name]; !ok {
					continue
				}
			}
			events = append(events, response.K8sEvent{
				Namespace:  event.Namespace,
				Type:       event.Type,
				Count:      event.Count,
				FirstOccur: event.FirstTimestamp.Unix(),
				LastSeen:   event.LastTimestamp.Unix(),
				Reason:     event.Reason,
				Message:    event.Message,
				Object:     fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			})
		}
	}
	return events, err
}

func ListK8sClusterNodes(ctx context.Context, c *client.Client) ([]response.K8sNode, error) {
	nodeList, err := c.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	nodes := make([]response.K8sNode, 0)
	nodeMetricsMap, metricsErr := resource.ListNodeMetrics(ctx)
	if err == nil {
		for _, node := range nodeList.Items {
			internalAddress, externalAddress := extractNodeAddress(&node)
			taints := make([]k8s.Taint, 0)
			for _, taint := range node.Spec.Taints {
				taints = append(taints, k8s.Taint{
					Key:    taint.Key,
					Value:  taint.Value,
					Effect: string(taint.Effect),
				})
			}
			nodeInfo := &response.K8sNodeInfo{
				Name:       node.Name,
				Status:     extractNodeStatus(&node),
				Roles:      extractNodeRoles(&node),
				Labels:     common.MapToKVs(node.Labels),
				Taints:     taints,
				Conditions: extractNodeConditions(&node),
				Uptime:     node.CreationTimestamp.Unix(),
				InternalIP: internalAddress,
				ExternalIP: externalAddress,
				Version:    node.Status.NodeInfo.KubeletVersion,
				OS:         node.Status.NodeInfo.OSImage,
				Kernel:     node.Status.NodeInfo.KernelVersion,
				CRI:        node.Status.NodeInfo.ContainerRuntimeVersion,
			}

			nodeResource := &response.K8sNodeResource{}
			if metricsErr == nil {
				nodeResource = extractNodeResource(nodeMetricsMap, &node)
			} else {
				logger.Errorf("Got error when list node metrics, err: %v", metricsErr)
			}
			nodes = append(nodes, response.K8sNode{
				Info:     nodeInfo,
				Resource: nodeResource,
			})
		}
	}
	return nodes, err
}
