package k8s

import (
	"fmt"
	"strings"

	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/oceanbase-dashboard/internal/business/common"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/constant"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/resource"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func CreateNamespace(param *param.CreateNamespaceParam) error {
	return resource.CreateNamespace(param.Namespace)
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

func extractNodeResource(node *corev1.Node) *response.K8sNodeResource {
	nodeResource := &response.K8sNodeResource{}
	nodeResource.CpuTotal = node.Status.Capacity.Cpu().AsApproximateFloat64()
	nodeResource.MemoryTotal = node.Status.Capacity.Memory().AsApproximateFloat64() / constant.GB
	podList, err := resource.ListAllPods()
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
					cpuRequested = cpuRequested + cpuRequest.AsApproximateFloat64()
				}
				memoryRequest, found := container.Resources.Requests[ResourceMemory]
				if found {
					memoryRequested = memoryRequested + memoryRequest.AsApproximateFloat64()/constant.GB
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

func ListEvents(queryEventParam *param.QueryEventParam) ([]response.K8sEvent, error) {
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
	if len(selectors) > 0 {
		listOptions.FieldSelector = strings.Join(selectors, ",")
	}
	eventList, err := resource.ListEvents(ns, listOptions)
	logger.Infof("query events with param: %v", queryEventParam)
	if err == nil {
		for _, event := range eventList.Items {
			events = append(events, response.K8sEvent{
				Namespace:  event.Namespace,
				Type:       event.Type,
				Count:      event.Count,
				FirstOccur: float64(event.FirstTimestamp.UnixMilli()) / 1000,
				LastSeen:   float64(event.LastTimestamp.UnixMilli()) / 1000,
				Reason:     event.Reason,
				Message:    event.Message,
				Object:     fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			})
		}
	}
	return events, err
}

func ListNodes() ([]response.K8sNode, error) {
	nodes := make([]response.K8sNode, 0)
	nodeList, err := resource.ListNodes()
	if err == nil {
		for _, node := range nodeList.Items {
			internalAddress, externalAddress := extractNodeAddress(&node)
			nodeInfo := &response.K8sNodeInfo{
				Name:       node.Name,
				Status:     extractNodeStatus(&node),
				Roles:      extractNodeRoles(&node),
				Labels:     common.MapToKVs(node.Labels),
				Conditions: extractNodeConditions(&node),
				Uptime:     float64(node.CreationTimestamp.UnixMilli()) / 1000,
				InternalIP: internalAddress,
				ExternalIP: externalAddress,
				Version:    node.Status.NodeInfo.KubeletVersion,
				OS:         node.Status.NodeInfo.OSImage,
				Kernel:     node.Status.NodeInfo.KernelVersion,
				CRI:        node.Status.NodeInfo.ContainerRuntimeVersion,
			}

			nodes = append(nodes, response.K8sNode{
				Info:     nodeInfo,
				Resource: extractNodeResource(&node),
			})
		}

	}
	return nodes, err
}

func ListStorageClasses() ([]response.StorageClass, error) {
	storageClasses := make([]response.StorageClass, 0)
	storageClassList, err := resource.ListStorageClasses()
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

func ListNamespaces() ([]response.Namespace, error) {
	namespaces := make([]response.Namespace, 0)
	namespaceList, err := resource.ListNamespaces()
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
