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

package obclusteroperation

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*OBClusterOperationManager]()

func ModifyClusterSpec(m *OBClusterOperationManager) tasktypes.TaskError {
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.OBCluster,
	}, obcluster)
	if err != nil {
		m.Logger.Error(err, "Failed to find obcluster")
		return err
	}
	origin := obcluster.DeepCopy()
	switch m.Resource.Spec.Type {
	case constants.ClusterOpTypeAddZones:
		if len(m.Resource.Spec.AddZones) == 0 {
			return errors.New("AddZones is empty")
		}
		obcluster.Spec.Topology = append(obcluster.Spec.Topology, m.Resource.Spec.AddZones...)
	case constants.ClusterOpTypeDeleteZones:
		if len(m.Resource.Spec.DeleteZones) == 0 {
			return errors.New("DeleteZones is empty")
		}
		deletingMap := make(map[string]struct{})
		for _, zone := range m.Resource.Spec.DeleteZones {
			deletingMap[zone] = struct{}{}
		}
		remainList := make([]apitypes.OBZoneTopology, 0)
		for i, t := range obcluster.Spec.Topology {
			if _, ok := deletingMap[t.Zone]; !ok {
				remainList = append(remainList, obcluster.Spec.Topology[i])
			}
		}
		obcluster.Spec.Topology = remainList
	case constants.ClusterOpTypeAdjustReplicas:
		if len(m.Resource.Spec.AdjustReplicas) == 0 {
			return errors.New("AdjustReplicas is empty")
		}
		for _, adjust := range m.Resource.Spec.AdjustReplicas {
			adjustingMap := make(map[string]struct{})
			for _, a := range adjust.Zones {
				adjustingMap[a] = struct{}{}
			}
			for i, t := range obcluster.Spec.Topology {
				if _, ok := adjustingMap[t.Zone]; ok {
					if adjust.To != 0 {
						obcluster.Spec.Topology[i].Replica = adjust.To
					} else {
						obcluster.Spec.Topology[i].Replica += adjust.By
					}
				}
			}
		}
	case constants.ClusterOpTypeRestartOBServers:
		if m.Resource.Spec.RestartOBServers == nil {
			return errors.New("RestartOBServers is empty")
		}
		if obcluster.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] != "true" {
			return errors.New("RestartOBServers only support static ip")
		}
		if m.Resource.Spec.RestartOBServers.Resource != nil {
			obcluster.Spec.OBServerTemplate.Resource = m.Resource.Spec.RestartOBServers.Resource
		}
		if m.Resource.Spec.RestartOBServers.AddingMonitor != nil {
			obcluster.Spec.MonitorTemplate = m.Resource.Spec.RestartOBServers.AddingMonitor
		}
		if m.Resource.Spec.RestartOBServers.AddingBackupVolume != nil {
			obcluster.Spec.BackupVolume = m.Resource.Spec.RestartOBServers.AddingBackupVolume
		}
	case constants.ClusterOpTypeUpgrade:
		if m.Resource.Spec.Upgrade == nil {
			return errors.New("Upgrade is empty")
		}
		if m.Resource.Spec.Upgrade.Image == "" {
			return errors.New("Upgrading image is empty")
		}
		obcluster.Spec.OBServerTemplate.Image = m.Resource.Spec.Upgrade.Image
	case constants.ClusterOpTypeExpandStorageSize:
		if m.Resource.Spec.ExpandStorageSize == nil {
			return errors.New("ModifyStorageSize is empty")
		}
		mutation := m.Resource.Spec.ExpandStorageSize
		if mutation.DataStorage != nil {
			obcluster.Spec.OBServerTemplate.Storage.DataStorage.Size = *mutation.DataStorage
		}
		if mutation.LogStorage != nil {
			obcluster.Spec.OBServerTemplate.Storage.LogStorage.Size = *mutation.LogStorage
		}
		if mutation.RedoLogStorage != nil {
			obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size = *mutation.RedoLogStorage
		}
	case constants.ClusterOpTypeModifyStorageClass:
		if m.Resource.Spec.ModifyStorageClass == nil {
			return errors.New("ModifyStorageClass is empty")
		}
		mutation := m.Resource.Spec.ModifyStorageClass
		if mutation.DataStorage != "" {
			obcluster.Spec.OBServerTemplate.Storage.DataStorage.StorageClass = mutation.DataStorage
		}
		if mutation.LogStorage != "" {
			obcluster.Spec.OBServerTemplate.Storage.LogStorage.StorageClass = mutation.LogStorage
		}
		if mutation.RedoLogStorage != "" {
			obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass = mutation.RedoLogStorage
		}
	case constants.ClusterOpTypeSetParameters:
		if m.Resource.Spec.SetParameters == nil {
			return errors.New("setParameters is empty")
		}
		newParamMap := make(map[string]string)
		for _, v := range m.Resource.Spec.SetParameters {
			newParamMap[v.Name] = v.Value
		}
		existingMap := make(map[string]struct{})
		for _, v := range obcluster.Spec.Parameters {
			if _, ok := newParamMap[v.Name]; ok {
				v.Value = newParamMap[v.Name]
			}
			existingMap[v.Name] = struct{}{}
		}
		for k, v := range newParamMap {
			if _, ok := existingMap[k]; !ok {
				obcluster.Spec.Parameters = append(obcluster.Spec.Parameters, apitypes.Parameter{Name: k, Value: v})
			}
		}
	}
	if m.Resource.Spec.Force {
		obcluster.Status.Status = clusterstatus.Running
		obcluster.Status.OperationContext = nil
	}
	err = m.Client.Patch(m.Ctx, obcluster, client.MergeFrom(origin))
	if err != nil {
		m.Logger.Error(err, "Failed to patch obcluster")
		return err
	}

	err = m.waitForOBClusterToBeStatus(obcfg.GetConfig().Time.DefaultStateWaitTimeout, func(status string) bool {
		return status != clusterstatus.Running
	})
	if err != nil {
		return errors.New("Timeout to wait for cluster to be operating")
	}
	return nil
}

func WaitForCluster(m *OBClusterOperationManager) tasktypes.TaskError {
	timeout := obcfg.GetConfig().Time.DefaultStateWaitTimeout
	if m.Resource.Spec.Type == constants.ClusterOpTypeModifyStorageClass {
		timeout = obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds
	}
	err := m.waitForOBClusterToBeStatus(timeout, func(status string) bool {
		return status == clusterstatus.Running
	})
	if err != nil {
		return errors.New("Timeout to wait for cluster to be running")
	}
	return nil
}

func RestartServers(m *OBClusterOperationManager) tasktypes.TaskError {
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.OBCluster,
	}, obcluster)
	if err != nil {
		m.Logger.Error(err, "Failed to find obcluster")
		return err
	}
	if obcluster.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] != "true" {
		return errors.New("RestartOBServers only support static ip")
	}

	observerList := v1alpha1.OBServerList{}
	err = m.Client.List(m.Ctx, &observerList, client.InNamespace(m.Resource.Namespace), client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: m.Resource.Spec.OBCluster,
	})
	if err != nil {
		m.Logger.Error(err, "Failed to list observers")
		return err
	}

	for _, observer := range observerList.Items {
		pod := corev1.Pod{}
		err = m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: observer.Namespace,
			Name:      observer.Name,
		}, &pod)
		if err != nil {
			m.Logger.Error(err, "Failed to find pod")
			return err
		}
		err = m.Client.Delete(m.Ctx, &pod)
		if err != nil {
			m.Logger.Error(err, "Failed to delete pod")
			return err
		}
		timeout := obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds
		err = m.waitForOBServerToBeStatus(observer.Name, timeout, func(status string) bool {
			return status != serverstatus.Running
		})
		if err != nil {
			return errors.New("Timeout to wait for server to be operating")
		}
		err = m.waitForOBServerToBeStatus(observer.Name, timeout, func(status string) bool {
			return status == serverstatus.Running
		})
		if err != nil {
			return errors.New("Timeout to wait for server to be running")
		}
	}
	return nil
}
