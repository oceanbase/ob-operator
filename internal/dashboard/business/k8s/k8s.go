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
	"errors"
	"fmt"
	"strings"

	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/obproxy"
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/k8s"
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

func extractNodeResource(metricsMap map[string]metricsv1beta1.NodeMetrics, node *corev1.Node) *response.K8sNodeResource {
	nodeResource := &response.K8sNodeResource{}
	metrics, ok := metricsMap[node.Name]
	nodeResource.CpuTotal = node.Status.Capacity.Cpu().AsApproximateFloat64()
	nodeResource.MemoryTotal = node.Status.Capacity.Memory().AsApproximateFloat64() / constant.GB
	if ok {
		if cpuUsed, ok := metrics.Usage[corev1.ResourceCPU]; ok {
			nodeResource.CpuUsed = cpuUsed.AsApproximateFloat64()
		}
		if memoryUsed, ok := metrics.Usage[corev1.ResourceMemory]; ok {
			nodeResource.MemoryUsed = memoryUsed.AsApproximateFloat64() / constant.GB
		}
		nodeResource.CpuFree = nodeResource.CpuTotal - nodeResource.CpuUsed
		nodeResource.MemoryFree = nodeResource.MemoryTotal - nodeResource.MemoryUsed
	}
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

func GetNode(ctx context.Context, name string) (*response.K8sNode, error) {
	node, err := resource.GetNode(ctx, name)
	if err != nil {
		return nil, err
	}
	return NewK8sNodeResponse(node), nil
}

func NewK8sNodeResponse(node *corev1.Node) *response.K8sNode {
	if node == nil {
		return nil
	}
	internalAddress, externalAddress := extractNodeAddress(node)
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
		Status:     extractNodeStatus(node),
		Roles:      extractNodeRoles(node),
		Labels:     common.MapToKVs(node.Labels),
		Taints:     taints,
		Conditions: extractNodeConditions(node),
		Uptime:     node.CreationTimestamp.Unix(),
		InternalIP: internalAddress,
		ExternalIP: externalAddress,
		Version:    node.Status.NodeInfo.KubeletVersion,
		OS:         node.Status.NodeInfo.OSImage,
		Kernel:     node.Status.NodeInfo.KernelVersion,
		CRI:        node.Status.NodeInfo.ContainerRuntimeVersion,
	}

	nodeResource := &response.K8sNodeResource{}
	nodeResp := &response.K8sNode{
		Info:     nodeInfo,
		Resource: nodeResource,
	}
	return nodeResp
}

func UpdateNodeTaints(ctx context.Context, name string, taints []k8s.Taint) (*response.K8sNode, error) {
	node, err := resource.GetNode(ctx, name)
	if err != nil {
		return nil, err
	} else {
		nodeTaints := make([]corev1.Taint, 0)
		for _, taint := range taints {
			nodeTaints = append(nodeTaints, corev1.Taint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: corev1.TaintEffect(taint.Effect),
			})
		}
		node.Spec.Taints = nodeTaints
		node, err = resource.UpdateNode(ctx, node)
		return NewK8sNodeResponse(node), err
	}
}

func UpdateNodeLabels(ctx context.Context, name string, labels []modelcommon.KVPair) (*response.K8sNode, error) {
	node, err := resource.GetNode(ctx, name)
	if err != nil {
		return nil, err
	} else {
		node.Labels = common.KVsToMap(labels)
		node, err = resource.UpdateNode(ctx, node)
		return NewK8sNodeResponse(node), err
	}
}

func BatchUpdateNodes(ctx context.Context, updateNodesParam *param.BatchUpdateNodesParam) error {
	failedNodes := make([]string, 0)
	for _, nodeName := range updateNodesParam.Nodes {
		node, err := resource.GetNode(ctx, nodeName)
		if err != nil {
			failedNodes = append(failedNodes, nodeName)
			continue
		}
		// update node labels
		for _, labelOperation := range updateNodesParam.LabelOperations {
			switch labelOperation.Operation {
			case param.OperationOverwrite:
				node.Labels[labelOperation.Key] = labelOperation.Value
			case param.OperationDelete:
				_, exists := node.Labels[labelOperation.Key]
				if exists {
					delete(node.Labels, labelOperation.Key)
				}
			default:
				logger.Errorf("Got unexpected node label operation: %s", labelOperation.Operation)
			}
		}
		// update node taints
		taintMap := make(map[string]*corev1.Taint)
		for idx, taint := range node.Spec.Taints {
			taintMap[taint.Key] = &node.Spec.Taints[idx]
		}
		for _, taintOperation := range updateNodesParam.TaintOperations {
			switch taintOperation.Operation {
			case param.OperationOverwrite:
				taintMap[taintOperation.Key] = &corev1.Taint{
					Key:    taintOperation.Key,
					Value:  taintOperation.Value,
					Effect: corev1.TaintEffect(taintOperation.Effect),
				}
			case param.OperationDelete:
				_, exists := taintMap[taintOperation.Key]
				if exists {
					delete(taintMap, taintOperation.Key)
				}
			default:
				logger.Errorf("Got unexpected node taint operation: %s", taintOperation.Operation)
			}
		}
		taints := make([]corev1.Taint, 0)
		for _, taint := range taintMap {
			taints = append(taints, *taint)
		}
		node.Spec.Taints = taints

		// update node
		node, err = resource.UpdateNode(ctx, node)
		if err != nil {
			failedNodes = append(failedNodes, nodeName)
			logger.Errorf("Got error when update node %s, %v", nodeName, err)
		}
	}
	if len(failedNodes) > 0 {
		return errors.New(fmt.Sprintf("Update nodes failed, failed nodes: %s", strings.Join(failedNodes, ",")))
	}
	return nil
}

func ListNodes(ctx context.Context) ([]response.K8sNode, error) {
	nodes := make([]response.K8sNode, 0)
	nodeList, err := resource.ListNodes(ctx)
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
