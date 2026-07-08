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
//go:generate task_register $GOFILE

package oblogservicezone

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	nodestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicenode"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskMap = builder.NewTaskHub[*OBLogServiceZoneManager]()

func CreateNodes(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Creating log service nodes")
	nodeList, err := m.listNodes()
	if err != nil {
		return errors.Wrap(err, "list existing nodes")
	}

	existingCount := 0
	for _, node := range nodeList.Items {
		if node.Status.Status != nodestatus.Unrecoverable {
			existingCount++
		}
	}
	desiredCount := m.Resource.Spec.Topology.Replica

	blockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		APIVersion:         m.Resource.APIVersion,
		Kind:               m.Resource.Kind,
		Name:               m.Resource.Name,
		UID:                m.Resource.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}

	for i := existingCount; i < desiredCount; i++ {
		nodeName := fmt.Sprintf("%s-%s", m.Resource.Name, rand.String(6))
		node := &v1alpha1.OBLogServiceNode{
			ObjectMeta: metav1.ObjectMeta{
				Name:            nodeName,
				Namespace:       m.Resource.Namespace,
				OwnerReferences: []metav1.OwnerReference{ownerRef},
				Labels: map[string]string{
					oceanbaseconst.LabelRefOBLogServiceCluster: m.Resource.Spec.ClusterName,
					oceanbaseconst.LabelRefOBLogServiceZone:    m.Resource.Name,
				},
			},
			Spec: v1alpha1.OBLogServiceNodeSpec{
				ClusterName: m.Resource.Spec.ClusterName,
				ClusterId:   m.Resource.Spec.ClusterId,
				Zone:        m.Resource.Spec.Topology.Zone,
				Image:       m.Resource.Spec.Image,
				Resource: func() *apitypes.ResourceSpec {
					if m.Resource.Spec.Topology.Resource != nil {
						return m.Resource.Spec.Topology.Resource
					}
					return &m.Resource.Spec.Resource
				}(),
				RpcPort:        m.Resource.Spec.Topology.RpcPort,
				HttpPort:       m.Resource.Spec.Topology.HttpPort,
				NodeSelector:   m.Resource.Spec.Topology.NodeSelector,
				Affinity:       m.Resource.Spec.Topology.Affinity,
				Tolerations:    m.Resource.Spec.Topology.Tolerations,
				ObjectStoreURL: m.Resource.Spec.ObjectStoreURL,
				Storage:        m.Resource.Spec.Storage,
				Parameters:     m.Resource.Spec.Parameters,
				ServiceAccount: m.Resource.Spec.ServiceAccount,
			},
		}
		modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.Resource, oceanbaseconst.AnnotationsMode)
		if modeAnnoExist {
			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
			}
			node.Annotations[oceanbaseconst.AnnotationsMode] = modeAnnoVal
		}
		if err := m.Client.Create(m.Ctx, node); err != nil {
			if !kubeerrors.IsAlreadyExists(err) {
				return errors.Wrapf(err, "create log service node %s", nodeName)
			}
		}
		m.Logger.Info("Created log service node", "name", nodeName)
	}
	return nil
}

func WaitNodesReady(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for log service nodes to be ready (bootstrap ready)")
	for i := 0; i < 600; i++ {
		nodeList, err := m.listNodes()
		if err != nil {
			return errors.Wrap(err, "list nodes")
		}
		if len(nodeList.Items) < m.Resource.Spec.Topology.Replica {
			time.Sleep(time.Second)
			continue
		}
		allReady := true
		for _, node := range nodeList.Items {
			if node.Status.Status != nodestatus.BootstrapReady && node.Status.Status != nodestatus.Running {
				allReady = false
				break
			}
		}
		if allReady {
			m.Logger.Info("All log service nodes are ready")
			nodeStatusList := make([]apitypes.LogServiceNodeReplicaStatus, 0, len(nodeList.Items))
			for _, node := range nodeList.Items {
				nodeStatusList = append(nodeStatusList, apitypes.LogServiceNodeReplicaStatus{
					NodeName: node.Name,
					Status:   node.Status.Status,
				})
			}
			m.Resource.Status.NodeStatus = nodeStatusList
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for log service nodes to be ready")
}

func WaitNodesRunning(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for log service nodes to be running")
	for i := 0; i < 600; i++ {
		nodeList, err := m.listNodes()
		if err != nil {
			return errors.Wrap(err, "list nodes")
		}
		runningCount := 0
		for _, node := range nodeList.Items {
			if node.Status.Status == nodestatus.Running {
				runningCount++
			}
		}
		if runningCount >= m.Resource.Spec.Topology.Replica {
			m.Logger.Info("All log service nodes are running")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for log service nodes to be running")
}

func DeleteExcessNodes(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Deleting excess log service nodes")
	nodeList, err := m.listNodes()
	if err != nil {
		return errors.Wrap(err, "list nodes")
	}

	desiredCount := m.Resource.Spec.Topology.Replica
	currentCount := len(nodeList.Items)
	if currentCount <= desiredCount {
		// No excess — clear any stale to-be-deleted marks (scale-up between
		// unregister and delete).
		for i := range nodeList.Items {
			node := &nodeList.Items[i]
			if node.Annotations != nil && node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted] == "true" {
				if err := m.clearToBeDeletedAnnotation(node); err != nil {
					return errors.Wrapf(err, "clear stale to-be-deleted mark on %s", node.Name)
				}
			}
		}
		return nil
	}

	toDelete := currentCount - desiredCount

	// Prefer nodes marked to-be-deleted by UnregisterNodeFromCluster so the
	// exact same set is deleted (the annotation is the source of truth).
	marked := make([]v1alpha1.OBLogServiceNode, 0, toDelete)
	unmarked := make([]v1alpha1.OBLogServiceNode, 0, len(nodeList.Items))
	for i := range nodeList.Items {
		if nodeList.Items[i].Annotations != nil && nodeList.Items[i].Annotations[oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted] == "true" {
			marked = append(marked, nodeList.Items[i])
		} else {
			unmarked = append(unmarked, nodeList.Items[i])
		}
	}

	// Build the deletion list: marked nodes first (up to toDelete), then
	// sorted fallback from unmarked if not enough marks exist.
	var candidates []v1alpha1.OBLogServiceNode
	if len(marked) >= toDelete {
		sortDeleteCandidates(marked)
		candidates = marked[:toDelete]
		// Clear marks on surplus marked nodes that won't be deleted.
		for i := toDelete; i < len(marked); i++ {
			if err := m.clearToBeDeletedAnnotation(&marked[i]); err != nil {
				return errors.Wrapf(err, "clear surplus mark on %s", marked[i].Name)
			}
		}
	} else {
		candidates = append(candidates, marked...)
		sortDeleteCandidates(unmarked)
		need := toDelete - len(marked)
		if need > len(unmarked) {
			need = len(unmarked)
		}
		candidates = append(candidates, unmarked[:need]...)
	}

	for i := range candidates {
		node := &candidates[i]
		if err := m.Client.Delete(m.Ctx, node); err != nil && !kubeerrors.IsNotFound(err) {
			return errors.Wrapf(err, "delete node %s", node.Name)
		}
		m.Logger.Info("Deleted log service node", "name", node.Name)
	}
	return nil
}

func DeleteAllNodes(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Deleting all log service nodes")
	nodeList, err := m.listNodes()
	if err != nil {
		return errors.Wrap(err, "list nodes")
	}
	for i := range nodeList.Items {
		if err := m.Client.Delete(m.Ctx, &nodeList.Items[i]); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return errors.Wrapf(err, "delete node %s", nodeList.Items[i].Name)
			}
		}
	}
	return nil
}

func RegisterNodeToCluster(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Registering new nodes to log service cluster")

	nodeList, err := m.listNodes()
	if err != nil {
		return errors.Wrap(err, "list nodes")
	}

	// Collect nodes that need registration
	var newNodes []*v1alpha1.OBLogServiceNode
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		if node.Status.Status != nodestatus.Running || node.Status.PodIP == "" {
			continue
		}
		if node.Annotations != nil && node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeRegistered] == "true" {
			continue
		}
		newNodes = append(newNodes, node)
	}
	if len(newNodes) == 0 {
		m.Logger.Info("No new nodes need registration")
		return nil
	}

	excludeNames := make([]string, 0, len(newNodes))
	for _, node := range newNodes {
		excludeNames = append(excludeNames, node.Name)
	}
	hostAddr, err := m.getRunningNodeHttpAddr(excludeNames...)
	if err != nil {
		return errors.Wrap(err, "get running node http addr")
	}

	var lastErr error
	for _, node := range newNodes {
		rpcPort := node.Spec.RpcPort
		if rpcPort == 0 {
			rpcPort = int32(oceanbaseconst.LogServiceRpcPort)
		}
		nodeAddr := fmt.Sprintf("%s:%d", node.Status.GetConnectAddr(), rpcPort)
		zoneName := node.Spec.Zone

		cmd := fmt.Sprintf(`/home/admin/oblogservice/bin/ls_ctrl --host %s add ln "%s" "%s"`,
			hostAddr, nodeAddr, zoneName)
		jobName := fmt.Sprintf("%s-add-ln", node.Name)

		m.Logger.Info("Running ls_ctrl add ln", "host", hostAddr, "node", nodeAddr, "zone", zoneName)
		_, exitCode, jobErr := resourceutils.RunJob(
			m.Ctx, m.Client, m.Logger, m.Resource.Namespace,
			jobName, m.Resource.Spec.Image, nil, cmd,
		)
		if jobErr != nil {
			m.Logger.Info("ls_ctrl add ln failed", "node", nodeAddr, "exitCode", exitCode, "err", jobErr)
			lastErr = jobErr
			continue
		}
		if patchErr := m.annotateNodeRegistered(node); patchErr != nil {
			return errors.Wrapf(patchErr, "annotate node %s as registered", node.Name)
		}
	}
	if lastErr != nil {
		return errors.Wrap(lastErr, "some nodes failed to register")
	}
	return nil
}

func (m *OBLogServiceZoneManager) annotateNodeRegistered(node *v1alpha1.OBLogServiceNode) error {
	patch := client.MergeFrom(node.DeepCopy())
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeRegistered] = "true"
	return m.Client.Patch(m.Ctx, node, patch)
}

func (m *OBLogServiceZoneManager) clearToBeDeletedAnnotation(node *v1alpha1.OBLogServiceNode) error {
	patch := client.MergeFrom(node.DeepCopy())
	delete(node.Annotations, oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted)
	return m.Client.Patch(m.Ctx, node, patch)
}

func UnregisterNodeFromCluster(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Unregistering nodes from log service cluster")

	nodeList, err := m.listNodes()
	if err != nil {
		return errors.Wrap(err, "list nodes")
	}

	desiredCount := m.Resource.Spec.Topology.Replica
	currentCount := len(nodeList.Items)
	if currentCount <= desiredCount {
		// Scale-up happened after we were scheduled — clear any stale marks.
		for i := range nodeList.Items {
			node := &nodeList.Items[i]
			if node.Annotations != nil && node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted] == "true" {
				if err := m.clearToBeDeletedAnnotation(node); err != nil {
					return errors.Wrapf(err, "clear stale to-be-deleted mark on %s", node.Name)
				}
			}
		}
		return nil
	}

	toDelete := currentCount - desiredCount

	// Prefer nodes already marked from a prior run (idempotent retry).
	alreadyMarked := make([]v1alpha1.OBLogServiceNode, 0)
	unmarked := make([]v1alpha1.OBLogServiceNode, 0)
	for i := range nodeList.Items {
		if nodeList.Items[i].Annotations != nil && nodeList.Items[i].Annotations[oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted] == "true" {
			alreadyMarked = append(alreadyMarked, nodeList.Items[i])
		} else {
			unmarked = append(unmarked, nodeList.Items[i])
		}
	}

	var toDeleteNodes []v1alpha1.OBLogServiceNode
	if len(alreadyMarked) >= toDelete {
		// Enough are already marked — use them and clear surplus marks.
		sortDeleteCandidates(alreadyMarked)
		toDeleteNodes = alreadyMarked[:toDelete]
		for i := toDelete; i < len(alreadyMarked); i++ {
			if err := m.clearToBeDeletedAnnotation(&alreadyMarked[i]); err != nil {
				return errors.Wrapf(err, "clear surplus mark on %s", alreadyMarked[i].Name)
			}
		}
	} else {
		// Need more candidates beyond what's already marked.
		toDeleteNodes = append(toDeleteNodes, alreadyMarked...)
		sortDeleteCandidates(unmarked)
		need := toDelete - len(alreadyMarked)
		if need > len(unmarked) {
			need = len(unmarked)
		}
		toDeleteNodes = append(toDeleteNodes, unmarked[:need]...)
	}

	// Persistently mark selected nodes so DeleteExcessNodes deletes the exact
	// same set even if the underlying list order changes between the two steps.
	for i := range toDeleteNodes {
		node := &toDeleteNodes[i]
		if node.Annotations != nil && node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted] == "true" {
			continue
		}
		patch := client.MergeFrom(node.DeepCopy())
		if node.Annotations == nil {
			node.Annotations = make(map[string]string)
		}
		node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeToBeDeleted] = "true"
		if err := m.Client.Patch(m.Ctx, node, patch); err != nil {
			return errors.Wrapf(err, "mark node %s to-be-deleted", node.Name)
		}
	}

	// Collect names of nodes to be deleted so we don't pick them as --host
	excludeNames := make([]string, 0, len(toDeleteNodes))
	for i := range toDeleteNodes {
		excludeNames = append(excludeNames, toDeleteNodes[i].Name)
	}
	hostAddr, err := m.getRunningNodeHttpAddr(excludeNames...)
	if err != nil {
		return errors.Wrap(err, "get running node http addr")
	}

	for i := range toDeleteNodes {
		node := &toDeleteNodes[i]
		nodeAddr := node.Status.GetConnectAddr()
		if nodeAddr == "" {
			m.Logger.Info("Node has no address, skip unregister", "node", node.Name)
			continue
		}
		rpcPort := node.Spec.RpcPort
		if rpcPort == 0 {
			rpcPort = int32(oceanbaseconst.LogServiceRpcPort)
		}
		nodeAddr = fmt.Sprintf("%s:%d", nodeAddr, rpcPort)

		cmd := fmt.Sprintf(`/home/admin/oblogservice/bin/ls_ctrl --host %s delete ln "%s"`,
			hostAddr, nodeAddr)
		jobName := fmt.Sprintf("%s-del-ln", node.Name)

		m.Logger.Info("Running ls_ctrl delete ln", "host", hostAddr, "node", nodeAddr)
		_, exitCode, jobErr := resourceutils.RunJob(
			m.Ctx, m.Client, m.Logger, m.Resource.Namespace,
			jobName, m.Resource.Spec.Image, nil, cmd,
		)
		if jobErr != nil {
			return errors.Wrapf(jobErr, "ls_ctrl delete ln failed for %s, exitCode=%d", nodeAddr, exitCode)
		}
	}
	return nil
}

func UnregisterAllNodesFromCluster(m *OBLogServiceZoneManager) tasktypes.TaskError {
	m.Logger.Info("Unregistering all nodes from log service cluster before zone deletion")
	hostAddr, err := m.getRunningNodeHttpAddr()
	if err != nil {
		// If no running node exists in cluster, zone may already be orphaned — skip
		m.Logger.Info("No running node found for unregister, skipping", "err", err)
		return nil
	}

	nodeList, err := m.listNodes()
	if err != nil {
		return errors.Wrap(err, "list nodes")
	}

	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		nodeAddr := node.Status.GetConnectAddr()
		if nodeAddr == "" {
			continue
		}
		rpcPort := node.Spec.RpcPort
		if rpcPort == 0 {
			rpcPort = int32(oceanbaseconst.LogServiceRpcPort)
		}
		nodeAddr = fmt.Sprintf("%s:%d", nodeAddr, rpcPort)

		cmd := fmt.Sprintf(`/home/admin/oblogservice/bin/ls_ctrl --host %s delete ln "%s"`,
			hostAddr, nodeAddr)
		jobName := fmt.Sprintf("%s-del-ln", node.Name)

		m.Logger.Info("Running ls_ctrl delete ln", "host", hostAddr, "node", nodeAddr)
		_, exitCode, jobErr := resourceutils.RunJob(
			m.Ctx, m.Client, m.Logger, m.Resource.Namespace,
			jobName, m.Resource.Spec.Image, nil, cmd,
		)
		if jobErr != nil {
			m.Logger.Info("ls_ctrl delete ln warning", "node", nodeAddr, "exitCode", exitCode, "err", jobErr)
		}
	}
	return nil
}

func (m *OBLogServiceZoneManager) getRunningNodeHttpAddr(excludeNames ...string) (string, error) {
	excludeSet := make(map[string]bool, len(excludeNames))
	for _, n := range excludeNames {
		excludeSet[n] = true
	}
	allNodes := &v1alpha1.OBLogServiceNodeList{}
	err := m.Client.List(m.Ctx, allNodes, client.MatchingLabels{
		oceanbaseconst.LabelRefOBLogServiceCluster: m.Resource.Spec.ClusterName,
	}, client.InNamespace(m.Resource.Namespace))
	if err != nil {
		return "", errors.Wrap(err, "list all cluster nodes")
	}
	for _, node := range allNodes.Items {
		if excludeSet[node.Name] {
			continue
		}
		if node.Status.Status == nodestatus.Running && node.Status.PodIP != "" {
			httpPort := node.Spec.HttpPort
			if httpPort == 0 {
				httpPort = int32(oceanbaseconst.LogServiceHttpPort)
			}
			return fmt.Sprintf("%s:%d", node.Status.GetConnectAddr(), httpPort), nil
		}
	}
	return "", errors.New("no running node with PodIP found in cluster")
}
