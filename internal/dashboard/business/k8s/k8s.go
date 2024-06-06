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

package k8s

import (
	"context"
	"fmt"
	"strings"

	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/obproxy"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/k8s/resource"
)

const (
	StatusReady    = "ready"
	StatusNotReady = "not ready"
)

const (
	ResourceCpu              = "cpu"
	ResourceMemory           = "memory"
	ResourceEphemeralStorage = "ephemeral-storage"
)

const (
	RoleLabelPrefix = "node-role.kubernetes.io"
)

func CreateNamespace(ctx context.Context, param *param.CreateNamespaceParam) error {
	return resource.CreateNamespace(ctx, param.Namespace)
}

func extractNodeStatus(node *corev1.Node) string {
	nodeStatus := StatusNotReady
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == corev1.ConditionTrue {
				nodeStatus = StatusReady
				break
			}
		}
	}
	return nodeStatus
}

func extractNodeRoles(node *corev1.Node) []string {
	roles := make([]string, 0)
	for key, value := range node.Labels {
		if strings.HasPrefix(key, RoleLabelPrefix) {
			labelParts := strings.Split(key, "/")
			if len(labelParts) >= 2 && value == "true" {
				roles = append(roles, labelParts[1])
			}
		}
	}
	return roles
}

func extractNodeAddress(node *corev1.Node) (string, string) {
	internalAddress := ""
	externalAddress := ""
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			internalAddress = address.Address
		}
		if address.Type == corev1.NodeExternalIP {
			externalAddress = address.Address
		}
	}
	return internalAddress, externalAddress
}

func extractNodeConditions(node *corev1.Node) []response.K8sNodeCondition {
	conditions := make([]response.K8sNodeCondition, 0)
	for _, condition := range node.Status.Conditions {
		if condition.Status == corev1.ConditionTrue && condition.Type != corev1.NodeReady {
			conditions = append(conditions, response.K8sNodeCondition{
				Type:    string(condition.Type),
				Reason:  condition.Reason,
				Message: condition.Message,
			})
		}
	}
	return conditions
}

func extractNodeResource(ctx context.Context, node *corev1.Node) *response.K8sNodeResource {
	nodeResource := &response.K8sNodeResource{}
	nodeResource.CpuTotal = node.Status.Capacity.Cpu().AsApproximateFloat64()
	nodeResource.MemoryTotal = node.Status.Capacity.Memory().AsApproximateFloat64() / constant.GB
	podList, err := resource.ListAllPods(ctx)
	if err == nil {
		cpuRequested := 0.0
		memoryRequested := 0.0
		for _, pod := range podList.Items {
			if !strings.Contains(pod.Spec.NodeName, node.Name) {
				continue
			}
			for _, container := range pod.Spec.Containers {
				cpuRequest, found := container.Resources.Requests[ResourceCpu]
				if found {
					cpuRequested += cpuRequest.AsApproximateFloat64()
				}
				memoryRequest, found := container.Resources.Requests[ResourceMemory]
				if found {
					memoryRequested += memoryRequest.AsApproximateFloat64() / constant.GB
				}
			}
		}
		nodeResource.CpuUsed = cpuRequested
		nodeResource.MemoryUsed = memoryRequested
	}
	nodeResource.CpuFree = nodeResource.CpuTotal - nodeResource.CpuUsed
	nodeResource.MemoryFree = nodeResource.MemoryTotal - nodeResource.MemoryUsed
	return nodeResource
}

func ListEvents(ctx context.Context, queryEventParam *param.QueryEventParam) ([]response.K8sEvent, error) {
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
		deployments, err := client.GetClient().MetaClient.Resource(schema.GroupVersionResource{
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

func ListNodes(ctx context.Context) ([]response.K8sNode, error) {
	nodes := make([]response.K8sNode, 0)
	nodeList, err := resource.ListNodes(ctx)
	if err == nil {
		for _, node := range nodeList.Items {
			internalAddress, externalAddress := extractNodeAddress(&node)
			nodeInfo := &response.K8sNodeInfo{
				Name:       node.Name,
				Status:     extractNodeStatus(&node),
				Roles:      extractNodeRoles(&node),
				Labels:     common.MapToKVs(node.Labels),
				Conditions: extractNodeConditions(&node),
				Uptime:     node.CreationTimestamp.Unix(),
				InternalIP: internalAddress,
				ExternalIP: externalAddress,
				Version:    node.Status.NodeInfo.KubeletVersion,
				OS:         node.Status.NodeInfo.OSImage,
				Kernel:     node.Status.NodeInfo.KernelVersion,
				CRI:        node.Status.NodeInfo.ContainerRuntimeVersion,
			}

			nodes = append(nodes, response.K8sNode{
				Info:     nodeInfo,
				Resource: extractNodeResource(ctx, &node),
			})
		}
	}
	return nodes, err
}

func ListNodeResources(ctx context.Context) ([]response.K8sNodeResource, error) {
	nodeList, err := resource.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	nodeResources := make([]response.K8sNodeResource, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		nodeResource := extractNodeResource(ctx, &node)
		nodeResources = append(nodeResources, *nodeResource)
	}
	return nodeResources, nil
}

func ListStorageClasses(ctx context.Context) ([]response.StorageClass, error) {
	storageClasses := make([]response.StorageClass, 0)
	storageClassList, err := resource.ListStorageClasses(ctx)
	if err == nil {
		for _, storageClass := range storageClassList.Items {
			volumeBindingMode := string(storagev1.VolumeBindingImmediate)
			if storageClass.VolumeBindingMode != nil {
				volumeBindingMode = string(*storageClass.VolumeBindingMode)
			}
			reclaimPolicy := string(corev1.PersistentVolumeReclaimDelete)
			if storageClass.ReclaimPolicy != nil {
				reclaimPolicy = string(*storageClass.ReclaimPolicy)
			}
			allowVolumeExpansion := false
			if storageClass.AllowVolumeExpansion != nil {
				allowVolumeExpansion = *storageClass.AllowVolumeExpansion
			}

			storageClasses = append(storageClasses, response.StorageClass{
				Name:                 storageClass.Name,
				Provisioner:          storageClass.Provisioner,
				ReclaimPolicy:        reclaimPolicy,
				VolumeBindingMode:    volumeBindingMode,
				Parameters:           common.MapToKVs(storageClass.Parameters),
				AllowVolumeExpansion: allowVolumeExpansion,
				MountOptions:         storageClass.MountOptions,
			})
		}
	}
	return storageClasses, err
}

func ListNamespaces(ctx context.Context) ([]response.Namespace, error) {
	namespaces := make([]response.Namespace, 0)
	namespaceList, err := resource.ListNamespaces(ctx)
	if err == nil {
		for _, namespace := range namespaceList.Items {
			namespaces = append(namespaces, response.Namespace{
				Namespace: namespace.Name,
				Status:    string(namespace.Status.Phase),
			})
		}
	}
	return namespaces, err
}
