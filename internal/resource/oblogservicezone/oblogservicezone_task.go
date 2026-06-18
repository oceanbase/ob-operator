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
	"sort"
	"time"

	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	nodestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicenode"
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
				Region:      m.Resource.Spec.Topology.Region,
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
			if node.Status.Status != nodestatus.Running {
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
		return nil
	}

	toDelete := currentCount - desiredCount
	// Sort: unrecoverable nodes first, then newest first
	sort.Slice(nodeList.Items, func(i, j int) bool {
		iUnrecoverable := nodeList.Items[i].Status.Status == nodestatus.Unrecoverable
		jUnrecoverable := nodeList.Items[j].Status.Status == nodestatus.Unrecoverable
		if iUnrecoverable != jUnrecoverable {
			return iUnrecoverable
		}
		return nodeList.Items[i].CreationTimestamp.After(nodeList.Items[j].CreationTimestamp.Time)
	})
	for i := 0; i < toDelete && i < len(nodeList.Items); i++ {
		node := &nodeList.Items[i]
		if err := m.Client.Delete(m.Ctx, node); err != nil {
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
