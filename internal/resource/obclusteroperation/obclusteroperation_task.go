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
	"strconv"

	"github.com/pkg/errors"
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
					if adjust.To > 0 {
						obcluster.Spec.Topology[i].Replica = adjust.To
					}
				}
			}
		}
	case constants.ClusterOpTypeRestartOBServers:
		// This is not a real operation, just a placeholder for the task
	case constants.ClusterOpTypeModifyOBServers:
		if m.Resource.Spec.ModifyOBServers == nil {
			return errors.New("modifyOBServers is empty")
		}
		if m.Resource.Spec.ModifyOBServers.ExpandStorageSize != nil {
			mutation := m.Resource.Spec.ModifyOBServers.ExpandStorageSize
			if mutation.DataStorage != nil {
				obcluster.Spec.OBServerTemplate.Storage.DataStorage.Size = *mutation.DataStorage
			}
			if mutation.LogStorage != nil {
				obcluster.Spec.OBServerTemplate.Storage.LogStorage.Size = *mutation.LogStorage
			}
			if mutation.RedoLogStorage != nil {
				obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size = *mutation.RedoLogStorage
			}
		}
		if m.Resource.Spec.ModifyOBServers.ModifyStorageClass != nil {
			mutation := m.Resource.Spec.ModifyOBServers.ModifyStorageClass
			if mutation.DataStorage != "" {
				obcluster.Spec.OBServerTemplate.Storage.DataStorage.StorageClass = mutation.DataStorage
			}
			if mutation.LogStorage != "" {
				obcluster.Spec.OBServerTemplate.Storage.LogStorage.StorageClass = mutation.LogStorage
			}
			if mutation.RedoLogStorage != "" {
				obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass = mutation.RedoLogStorage
			}
		}
		supportStaticIP := obcluster.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] == "true"
		if m.Resource.Spec.ModifyOBServers.AddingMonitor != nil && supportStaticIP {
			obcluster.Spec.MonitorTemplate = m.Resource.Spec.ModifyOBServers.AddingMonitor
		}
		if m.Resource.Spec.ModifyOBServers.AddingBackupVolume != nil && supportStaticIP {
			obcluster.Spec.BackupVolume = m.Resource.Spec.ModifyOBServers.AddingBackupVolume
		}
		if m.Resource.Spec.ModifyOBServers.RemoveBackupVolume && supportStaticIP {
			obcluster.Spec.BackupVolume = nil
		}
		if m.Resource.Spec.ModifyOBServers.RemoveMonitor && supportStaticIP {
			obcluster.Spec.MonitorTemplate = nil
		}
		if m.Resource.Spec.ModifyOBServers.Resource != nil && supportStaticIP {
			obcluster.Spec.OBServerTemplate.Resource = m.Resource.Spec.ModifyOBServers.Resource
		}
	case constants.ClusterOpTypeUpgrade:
		if m.Resource.Spec.Upgrade == nil {
			return errors.New("Upgrade is empty")
		}
		if m.Resource.Spec.Upgrade.Image == "" {
			return errors.New("Upgrading image is empty")
		}
		obcluster.Spec.OBServerTemplate.Image = m.Resource.Spec.Upgrade.Image
	case constants.ClusterOpTypeSetParameters:
		if m.Resource.Spec.SetParameters == nil {
			return errors.New("setParameters is empty")
		}
		newParamMap := make(map[string]string)
		for _, v := range m.Resource.Spec.SetParameters {
			newParamMap[v.Name] = v.Value
		}
		existingMap := make(map[string]struct{})
		for i, v := range obcluster.Spec.Parameters {
			if val, ok := newParamMap[v.Name]; ok {
				obcluster.Spec.Parameters[i].Value = val
			}
			existingMap[v.Name] = struct{}{}
		}
		for k, v := range newParamMap {
			if _, ok := existingMap[k]; !ok {
				obcluster.Spec.Parameters = append(obcluster.Spec.Parameters, apitypes.Parameter{Name: k, Value: v})
			}
		}
	case constants.ClusterOpTypeDeleteOBServers:
		if len(m.Resource.Spec.DeleteOBServers.OBServers) == 0 {
			return errors.New("Delete observers is empty")
		}
		observerList := v1alpha1.OBServerList{}
		err = m.Client.List(m.Ctx, &observerList, client.InNamespace(m.Resource.Namespace), client.MatchingLabels{
			oceanbaseconst.LabelRefOBCluster: m.Resource.Spec.OBCluster,
		})
		if err != nil {
			m.Logger.Error(err, "Failed to list observers")
			return err
		}
		zoneModificationMap := make(map[string]int)
		for _, observer := range observerList.Items {
			for _, observerName := range m.Resource.Spec.DeleteOBServers.OBServers {
				if observer.Name == observerName {
					val, ok := zoneModificationMap[observer.Spec.Zone]
					if !ok {
						val = 0
					}
					zoneModificationMap[observer.Spec.Zone] = val + 1
				}
			}
		}
		for i, t := range obcluster.Spec.Topology {
			if v, ok := zoneModificationMap[t.Zone]; ok {
				if t.Replica-v > 0 {
					obcluster.Spec.Topology[i].Replica = t.Replica - v
				}
			}
		}
	}
	if m.Resource.Spec.Force {
		obcluster.Status.Status = clusterstatus.Running
		obcluster.Status.OperationContext = nil
	} else if obcluster.Status.Status != clusterstatus.Running {
		return errors.New("obcluster is not running")
	}
	oldResourceVersion := obcluster.ResourceVersion
	err = m.Client.Patch(m.Ctx, obcluster, client.MergeFrom(origin))
	if err != nil {
		m.Logger.Error(err, "Failed to patch obcluster")
		return err
	}
	newResourceVersion := obcluster.ResourceVersion
	if oldResourceVersion == newResourceVersion {
		m.Logger.Info("obcluster not changed")
		return nil
	}
	err = m.waitForOBClusterStatusToMatch(obcfg.GetConfig().Time.DefaultStateWaitTimeout, func(status string) bool {
		return status != clusterstatus.Running
	})
	if err != nil {
		return errors.New("Timeout to wait for cluster to be operating")
	}
	return nil
}

func WaitForClusterReturnRunning(m *OBClusterOperationManager) tasktypes.TaskError {
	timeout := obcfg.GetConfig().Time.DefaultStateWaitTimeout
	if m.Resource.Spec.Type == constants.ClusterOpTypeModifyOBServers &&
		m.Resource.Spec.ModifyOBServers != nil &&
		m.Resource.Spec.ModifyOBServers.ModifyStorageClass != nil {
		timeout = obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds
	}
	err := m.waitForOBClusterStatusToMatch(timeout, func(status string) bool {
		return status == clusterstatus.Running
	})
	if err != nil {
		return errors.New("Timeout to wait for cluster to be running")
	}
	return nil
}

func AnnotateOBServersForDeletion(m *OBClusterOperationManager) tasktypes.TaskError {
	var err error
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.OBCluster,
	}, obcluster)
	if err != nil {
		m.Logger.Error(err, "Failed to find obcluster")
		return err
	}
	if obcluster.Status.Status != clusterstatus.Running && !m.Resource.Spec.Force {
		return errors.New("RestartOBServers requires obcluster to be running")
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
		for _, observerName := range m.Resource.Spec.DeleteOBServers.OBServers {
			if observer.Name == observerName {
				if observer.Annotations == nil {
					observer.Annotations = make(map[string]string)
				}
				observer.Annotations[oceanbaseconst.AnnotationsDeletionPriority] = strconv.Itoa(oceanbaseconst.DefaultDeletionPriority)
				err = m.Client.Update(m.Ctx, &observer)
				if err != nil {
					return errors.Wrapf(err, "Failed to annotate observer %s", observerName)
				}
			}
		}
	}
	return nil
}

func RestartOBServers(m *OBClusterOperationManager) tasktypes.TaskError {
	restartingServers := make([]v1alpha1.OBServer, 0)
	var err error
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.OBCluster,
	}, obcluster)
	if err != nil {
		m.Logger.Error(err, "Failed to find obcluster")
		return err
	}
	if obcluster.Status.Status != clusterstatus.Running && !m.Resource.Spec.Force {
		return errors.New("RestartOBServers requires obcluster to be running")
	}
	if obcluster.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] != "true" {
		return errors.New("RestartOBServers requires obcluster's support for static ip")
	}

	observerList := v1alpha1.OBServerList{}
	err = m.Client.List(m.Ctx, &observerList, client.InNamespace(m.Resource.Namespace), client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: m.Resource.Spec.OBCluster,
	})
	if err != nil {
		m.Logger.Error(err, "Failed to list observers")
		return err
	}

	if m.Resource.Spec.RestartOBServers.All {
		restartingServers = append(restartingServers, observerList.Items...)
	} else if len(m.Resource.Spec.RestartOBServers.OBZones) > 0 {
		filterZoneMap := make(map[string]struct{})
		for _, zone := range m.Resource.Spec.RestartOBServers.OBZones {
			filterZoneMap[zone] = struct{}{}
		}
		for _, observer := range observerList.Items {
			if _, ok := filterZoneMap[observer.Labels[oceanbaseconst.LabelRefOBZone]]; ok {
				restartingServers = append(restartingServers, observer)
			}
		}
	} else if len(m.Resource.Spec.RestartOBServers.OBServers) > 0 {
		filterObserverMap := make(map[string]struct{})
		for _, observer := range m.Resource.Spec.RestartOBServers.OBServers {
			filterObserverMap[observer] = struct{}{}
		}
		for _, observer := range observerList.Items {
			if _, ok := filterObserverMap[observer.Name]; ok {
				restartingServers = append(restartingServers, observer)
			}
		}
	}

	for _, observer := range restartingServers {
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
		err = m.waitForOBServerStatusToMatch(observer.Name, timeout, func(status string) bool {
			return status != serverstatus.Running
		})
		if err != nil {
			return errors.New("Timeout to wait for server to be operating")
		}
		err = m.waitForOBServerStatusToMatch(observer.Name, timeout, func(status string) bool {
			return status == serverstatus.Running
		})
		if err != nil {
			return errors.New("Timeout to wait for server to be running")
		}
	}
	return nil
}
