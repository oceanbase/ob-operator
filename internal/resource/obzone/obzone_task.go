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

package obzone

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task-register $GOFILE

var taskMap = builder.NewTaskHub[*OBZoneManager]()

func AddZone(m *OBZoneManager) tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	return oceanbaseOperationManager.AddZone(m.Ctx, m.OBZone.Spec.Topology.Zone)
}

func StartOBZone(m *OBZoneManager) tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	return oceanbaseOperationManager.StartZone(m.Ctx, m.OBZone.Spec.Topology.Zone)
}

func CreateOBServer(m *OBZoneManager) tasktypes.TaskError {
	m.Logger.Info("Create observers")
	blockOwnerDeletion := true
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion:         m.OBZone.APIVersion,
		Kind:               m.OBZone.Kind,
		Name:               m.OBZone.Name,
		UID:                m.OBZone.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	currentReplica := 0
	for _, observerStatus := range m.OBZone.Status.OBServerStatus {
		if observerStatus.Status != serverstatus.Unrecoverable {
			currentReplica++
		}
	}
	independentVolumeAnnoVal, independentVolumeAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsIndependentPVCLifecycle)
	singlePVCAnnoVal, singlePVCAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsSinglePVC)
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsMode)
	migrateAnnoVal, migrateAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsSourceClusterAddress)
	for i := currentReplica; i < m.OBZone.Spec.Topology.Replica; i++ {
		serverName := m.generateServerName()
		finalizerName := oceanbaseconst.FinalizerDeleteOBServer
		finalizers := []string{finalizerName}
		labels := make(map[string]string)
		cluster, _ := m.OBZone.Labels[oceanbaseconst.LabelRefOBCluster]
		labels[oceanbaseconst.LabelRefUID] = string(m.OBZone.GetUID())
		labels[oceanbaseconst.LabelRefOBZone] = m.OBZone.Name
		labels[oceanbaseconst.LabelRefOBCluster] = cluster
		observer := &v1alpha1.OBServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:            serverName,
				Namespace:       m.OBZone.Namespace,
				OwnerReferences: ownerReferenceList,
				Finalizers:      finalizers,
				Labels:          labels,
			},
			Spec: v1alpha1.OBServerSpec{
				ClusterName:      m.OBZone.Spec.ClusterName,
				ClusterId:        m.OBZone.Spec.ClusterId,
				Zone:             m.OBZone.Spec.Topology.Zone,
				NodeSelector:     m.OBZone.Spec.Topology.NodeSelector,
				Affinity:         m.OBZone.Spec.Topology.Affinity,
				Tolerations:      m.OBZone.Spec.Topology.Tolerations,
				OBServerTemplate: m.OBZone.Spec.OBServerTemplate,
				MonitorTemplate:  m.OBZone.Spec.MonitorTemplate,
				BackupVolume:     m.OBZone.Spec.BackupVolume,
				ServiceAccount:   m.OBZone.Spec.ServiceAccount,
			},
		}
		observer.ObjectMeta.Annotations = make(map[string]string)
		if independentVolumeAnnoExist {
			observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsIndependentPVCLifecycle] = independentVolumeAnnoVal
		}
		if singlePVCAnnoExist {
			observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSinglePVC] = singlePVCAnnoVal
		}
		if modeAnnoExist {
			observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsMode] = modeAnnoVal
		}
		if migrateAnnoExist {
			observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSourceClusterAddress] = migrateAnnoVal
		}
		m.Logger.Info("Create observer", "server", serverName)
		err := m.Client.Create(m.Ctx, observer)
		if err != nil {
			m.Logger.Error(err, "Create observer failed", "server", serverName)
			return errors.Wrap(err, "create observer")
		}
		m.Recorder.Event(m.OBZone, "CreateObServer", "CreateObserver", fmt.Sprintf("Create observer %s", serverName))
	}
	return nil
}

func DeleteOBServer(m *OBZoneManager) tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Delete observers")
	observerList, err := m.listOBServers()
	if err != nil {
		m.Logger.Error(err, "List observers failed")
		return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
	}
	observerCount := 0
	for _, observer := range observerList.Items {
		// bugfix: if an observer is being deleted, it won't be deleted again or counted as a working observer
		if observer.Status.Status == serverstatus.Deleting {
			continue
		}
		if observer.Status.Status == serverstatus.Unrecoverable || observerCount >= m.OBZone.Spec.Topology.Replica {
			m.Logger.Info("Delete observer", "observer", observer)
			err = m.Client.Delete(m.Ctx, &observer)
			if err != nil {
				return errors.Wrapf(err, "Delete observer %s failed", observer.Name)
			}
			m.Recorder.Event(m.OBZone, "DeleteObServer", "DeleteObserver", fmt.Sprintf("Delete observer %+v", observer))
			continue
		}
		observerCount++
	}
	return nil
}

// TODO refactor Delete observer method together
func DeleteAllOBServer(m *OBZoneManager) tasktypes.TaskError {
	m.Logger.Info("Delete all observers")
	observerList, err := m.listOBServers()
	if err != nil {
		m.Logger.Error(err, "List observers failed")
		return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
	}
	for _, observer := range observerList.Items {
		m.Logger.Info("Need to delete observer", "observer", observer.Name)
		err = m.Client.Delete(m.Ctx, &observer)
		if err != nil {
			return errors.Wrapf(err, "Delete observer %s failed", observer.Name)
		}
	}
	return nil
}

func WaitReplicaMatch(m *OBZoneManager) tasktypes.TaskError {
	matched := false
	for i := 0; i < obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds; i++ {
		obzone, err := m.getOBZone()
		if err != nil {
			m.Logger.Error(err, "Get obzone from K8s failed")
			return nil
		} else if m.OBZone.Spec.Topology.Replica == len(obzone.Status.OBServerStatus) {
			m.Logger.Info("OBZone replica matched")
			matched = true
			break
		} else {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Zone replica not match", "desired replica", m.OBZone.Spec.Topology.Replica, "current replica", len(m.OBZone.Status.OBServerStatus))
		}
		time.Sleep(time.Second * 1)
	}
	if !matched {
		return errors.Errorf("wait obzone %s replica match timeout", m.OBZone.Name)
	}
	return nil
}

func WaitOBServerDeleted(m *OBZoneManager) tasktypes.TaskError {
	matched := false
	for i := 0; i < obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds; i++ {
		obzone, err := m.getOBZone()
		if err != nil {
			m.Logger.Error(err, "Get obzone from K8s failed")
		}
		if 0 == len(obzone.Status.OBServerStatus) {
			m.Logger.Info("OBServer all deleted")
			matched = true
			break
		}
		time.Sleep(time.Second * 1)
	}
	if !matched {
		return errors.Errorf("wait obzone %s observer deleted timeout", m.OBZone.Name)
	}
	return nil
}

func StopOBZone(m *OBZoneManager) tasktypes.TaskError {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "OBZone %s get oceanbase operation manager", m.OBZone.Name)
	}
	err = operationManager.StopZone(m.Ctx, m.OBZone.Spec.Topology.Zone)
	if err != nil {
		return errors.Wrapf(err, "Stop obzone %s failed", m.OBZone.Spec.Topology.Zone)
	}
	return nil
}

func OBClusterHealthCheck(m *OBZoneManager) tasktypes.TaskError {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	_ = resourceutils.ExecuteUpgradeScript(m.Ctx, m.Client, m.Logger, obcluster, oceanbaseconst.UpgradeHealthCheckerScriptPath, "")
	return nil
}

func OBZoneHealthCheck(m *OBZoneManager) tasktypes.TaskError {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	zoneOpt := fmt.Sprintf("-z '%s'", m.OBZone.Spec.Topology.Zone)
	_ = resourceutils.ExecuteUpgradeScript(m.Ctx, m.Client, m.Logger, obcluster, oceanbaseconst.UpgradeHealthCheckerScriptPath, zoneOpt)
	return nil
}

func UpgradeOBServer(m *OBZoneManager) tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		observerList, err := m.listOBServers()
		if err != nil {
			m.Logger.Error(err, "List observers failed")
			return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
		}
		for _, observer := range observerList.Items {
			m.Logger.Info("Upgrade observer", "observer", observer.Name)
			observer.Spec.OBServerTemplate.Image = m.OBZone.Spec.OBServerTemplate.Image
			err = m.Client.Update(m.Ctx, &observer)
			if err != nil {
				return errors.Wrapf(err, "Upgrade observer %s failed", observer.Name)
			}
		}
		return nil
	})
}

func WaitOBServerUpgraded(m *OBZoneManager) tasktypes.TaskError {
	for i := 0; i < obcfg.GetConfig().Time.TimeConsumingStateWaitTimeout; i++ {
		observerList, err := m.listOBServers()
		if err != nil {
			m.Logger.Error(err, "List observers failed")
			return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
		}
		allServerUpgraded := true
		for _, observer := range observerList.Items {
			if !(observer.Status.Status == serverstatus.Running && observer.Status.Image == m.OBZone.Spec.OBServerTemplate.Image) {
				m.Logger.Info("Found observer upgrade not finished", "observer", observer.Name)
				allServerUpgraded = false
				break
			}
		}
		if allServerUpgraded {
			m.Logger.Info("All server upgraded")
			return nil
		}
		time.Sleep(time.Duration(obcfg.GetConfig().Time.CommonCheckInterval) * time.Second)
	}
	return errors.New("Wait all server upgraded timeout")
}

func DeleteOBZoneInCluster(m *OBZoneManager) tasktypes.TaskError {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "OBZone %s get oceanbase operation manager", m.OBZone.Name)
	}
	err = operationManager.DeleteZone(m.Ctx, m.OBZone.Spec.Topology.Zone)
	if err != nil {
		return errors.Wrapf(err, "Delete obzone %s failed", m.OBZone.Spec.Topology.Zone)
	}
	return nil
}

func ScaleUpOBServers(m *OBZoneManager) tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	zoneRes := m.OBZone.Spec.OBServerTemplate.Resource
	for _, observer := range observerList.Items {
		serverRes := observer.Spec.OBServerTemplate.Resource
		if serverRes.Cpu != zoneRes.Cpu || serverRes.Memory != zoneRes.Memory {
			m.Logger.Info("Scale up observer", "observer", observer.Name)
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				serverRes.Cpu = zoneRes.Cpu
				serverRes.Memory = zoneRes.Memory
				return m.Client.Update(m.Ctx, &observer)
			})
			if err != nil {
				return errors.Wrapf(err, "Scale up observer %s failed", observer.Name)
			}
		}
	}
	return nil
}

func ExpandPVC(m *OBZoneManager) tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	zoneStorage := m.OBZone.Spec.OBServerTemplate.Storage
	for _, observer := range observerList.Items {
		serverStorage := observer.Spec.OBServerTemplate.Storage
		if serverStorage.DataStorage.Size.Cmp(zoneStorage.DataStorage.Size) < 0 ||
			serverStorage.LogStorage.Size.Cmp(zoneStorage.LogStorage.Size) < 0 ||
			serverStorage.RedoLogStorage.Size.Cmp(zoneStorage.RedoLogStorage.Size) < 0 {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				m.Logger.Info("Expand pvc of observer", "observer", observer.Name)
				serverStorage.DataStorage.Size = zoneStorage.DataStorage.Size
				serverStorage.LogStorage.Size = zoneStorage.LogStorage.Size
				serverStorage.RedoLogStorage.Size = zoneStorage.RedoLogStorage.Size
				return m.Client.Update(m.Ctx, &observer)
			})
			if err != nil {
				return errors.Wrapf(err, "Expand observer %s failed", observer.Name)
			}
		}
	}
	return nil
}

func MountBackupVolume(m *OBZoneManager) tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	for _, observer := range observerList.Items {
		if observer.Spec.BackupVolume == nil && m.OBZone.Spec.BackupVolume != nil {
			m.Logger.Info("Mount backup volume", "observer", observer.Name)
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				observer.Spec.BackupVolume = m.OBZone.Spec.BackupVolume
				return m.Client.Update(m.Ctx, &observer)
			})
			if err != nil {
				return errors.Wrapf(err, "Mount backup volume %s failed", observer.Name)
			}
		}
	}
	return nil
}

func DeleteLegacyOBServers(m *OBZoneManager) tasktypes.TaskError {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "OBZone %s get oceanbase operation manager", m.OBZone.Name)
	}
	allOBServers, err := operationManager.ListServers(m.Ctx)
	if err != nil {
		return errors.Wrap(err, "List observers in oceanbase")
	}
	observerList, err := m.listOBServers()
	if err != nil {
		return errors.Wrap(err, "List observer crs")
	}
	for _, observer := range allOBServers {
		// skip observers in different obzone
		if observer.Zone != m.OBZone.Spec.Topology.Zone {
			continue
		}
		found := false
		for _, observerCR := range observerList.Items {
			if observer.Ip == observerCR.Status.GetConnectAddr() {
				found = true
			}
		}
		if !found {
			err := operationManager.DeleteServer(m.Ctx, &model.ServerInfo{
				Ip:   observer.Ip,
				Port: observer.Port,
			})
			if err != nil {
				return errors.Wrapf(err, "Delete observer %s:%d", observer.Ip, observer.Port)
			}
		}
	}
	return nil
}

func WaitOBServerBootstrapReady(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.BootstrapReady, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitOBServerRunning(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.Running, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerScalingUp(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.ScaleUp, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerExpandingPVC(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.ExpandPVC, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerMounting(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.MountBackupVolume, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}
