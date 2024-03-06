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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func (m *OBZoneManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
}

func (m *OBZoneManager) generateServerName() string {
	parts := strings.Split(uuid.New().String(), "-")
	suffix := parts[len(parts)-1]
	return fmt.Sprintf("%s-%d-%s-%s", m.OBZone.Spec.ClusterName, m.OBZone.Spec.ClusterId, m.OBZone.Spec.Topology.Zone, suffix)
}

func (m *OBZoneManager) AddZone() tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	return oceanbaseOperationManager.AddZone(m.OBZone.Spec.Topology.Zone)
}

func (m *OBZoneManager) StartOBZone() tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	return oceanbaseOperationManager.StartZone(m.OBZone.Spec.Topology.Zone)
}

func (m *OBZoneManager) generateWaitOBServerStatusFunc(status string, timeoutSeconds int) tasktypes.TaskFunc {
	f := func() tasktypes.TaskError {
		for i := 1; i < timeoutSeconds; i++ {
			obzone, err := m.getOBZone()
			if err != nil {
				return errors.Wrap(err, "get obzoen failed")
			}
			allMatched := true
			for _, observerStatus := range obzone.Status.OBServerStatus {
				if observerStatus.Status != status && observerStatus.Status != serverstatus.Unrecoverable {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("server status still not matched", "server", observerStatus.Server, "status", status)
					allMatched = false
					break
				}
			}
			if allMatched {
				return nil
			}
			time.Sleep(time.Second)
		}
		return errors.New("all server still not bootstrap ready when timeout")
	}
	return f
}

func (m *OBZoneManager) CreateOBServer() tasktypes.TaskError {
	m.Logger.Info("create observers")
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
	for i := currentReplica; i < m.OBZone.Spec.Topology.Replica; i++ {
		serverName := m.generateServerName()
		finalizerName := "finalizers.oceanbase.com.deleteobserver"
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
		m.Logger.Info("create observer", "server", serverName)
		err := m.Client.Create(m.Ctx, observer)
		if err != nil {
			m.Logger.Error(err, "create observer failed", "server", serverName)
			return errors.Wrap(err, "create observer")
		}
		m.Recorder.Event(m.OBZone, "CreateObServer", "CreateObserver", fmt.Sprintf("Create observer %s", serverName))
	}
	return nil
}

func (m *OBZoneManager) DeleteOBServer() tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("delete observers")
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
			m.Logger.Info("delete observer", "observer", observer)
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
func (m *OBZoneManager) DeleteAllOBServer() tasktypes.TaskError {
	m.Logger.Info("delete all observers")
	observerList, err := m.listOBServers()
	if err != nil {
		m.Logger.Error(err, "List observers failed")
		return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
	}
	for _, observer := range observerList.Items {
		m.Logger.Info("need delete observer", "observer", observer.Name)
		err = m.Client.Delete(m.Ctx, &observer)
		if err != nil {
			return errors.Wrapf(err, "Delete observer %s failed", observer.Name)
		}
	}
	return nil
}

func (m *OBZoneManager) WaitReplicaMatch() tasktypes.TaskError {
	matched := false
	for i := 0; i < oceanbaseconst.ServerDeleteTimeoutSeconds; i++ {
		obzone, err := m.getOBZone()
		if err != nil {
			m.Logger.Error(err, "Get obzone from K8s failed")
		} else if m.OBZone.Spec.Topology.Replica == len(obzone.Status.OBServerStatus) {
			m.Logger.Info("Obzone replica matched")
			matched = true
			break
		} else {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("zone replica not match", "desired replica", m.OBZone.Spec.Topology.Replica, "current replica", len(m.OBZone.Status.OBServerStatus))
		}
		time.Sleep(time.Second * 1)
	}
	if !matched {
		return errors.Errorf("wait obzone %s replica match timeout", m.OBZone.Name)
	}
	return nil
}

func (m *OBZoneManager) WaitOBServerDeleted() tasktypes.TaskError {
	matched := false
	for i := 0; i < oceanbaseconst.ServerDeleteTimeoutSeconds; i++ {
		obzone, err := m.getOBZone()
		if err != nil {
			m.Logger.Error(err, "Get obzone from K8s failed")
		}
		if 0 == len(obzone.Status.OBServerStatus) {
			m.Logger.Info("observer all deleted")
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

func (m *OBZoneManager) StopOBZone() tasktypes.TaskError {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "OBZone %s get oceanbase operation manager", m.OBZone.Name)
	}
	err = operationManager.StopZone(m.OBZone.Spec.Topology.Zone)
	if err != nil {
		return errors.Wrapf(err, "Stop obzone %s failed", m.OBZone.Spec.Topology.Zone)
	}
	return nil
}

func (m *OBZoneManager) OBClusterHealthCheck() tasktypes.TaskError {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	_ = resourceutils.ExecuteUpgradeScript(m.Client, m.Logger, obcluster, oceanbaseconst.UpgradeHealthCheckerScriptPath, "")
	return nil
}

func (m *OBZoneManager) OBZoneHealthCheck() tasktypes.TaskError {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	zoneOpt := fmt.Sprintf("-z '%s'", m.OBZone.Spec.Topology.Zone)
	_ = resourceutils.ExecuteUpgradeScript(m.Client, m.Logger, obcluster, oceanbaseconst.UpgradeHealthCheckerScriptPath, zoneOpt)
	return nil
}

func (m *OBZoneManager) UpgradeOBServer() tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		observerList, err := m.listOBServers()
		if err != nil {
			m.Logger.Error(err, "List observers failed")
			return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
		}
		for _, observer := range observerList.Items {
			m.Logger.Info("upgrade observer", "observer", observer.Name)
			observer.Spec.OBServerTemplate.Image = m.OBZone.Spec.OBServerTemplate.Image
			err = m.Client.Update(m.Ctx, &observer)
			if err != nil {
				return errors.Wrapf(err, "Upgrade observer %s failed", observer.Name)
			}
		}
		return nil
	})
}

func (m *OBZoneManager) WaitOBServerUpgraded() tasktypes.TaskError {
	for i := 0; i < oceanbaseconst.TimeConsumingStateWaitTimeout; i++ {
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
		time.Sleep(oceanbaseconst.CommonCheckInterval * time.Second)
	}
	return errors.New("Wait all server upgraded timeout")
}

func (m *OBZoneManager) DeleteOBZoneInCluster() tasktypes.TaskError {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "OBZone %s get oceanbase operation manager", m.OBZone.Name)
	}
	err = operationManager.DeleteZone(m.OBZone.Spec.Topology.Zone)
	if err != nil {
		return errors.Wrapf(err, "Delete obzone %s failed", m.OBZone.Spec.Topology.Zone)
	}
	return nil
}

func (m *OBZoneManager) ScaleUpOBServer() tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	for _, observer := range observerList.Items {
		if observer.Spec.OBServerTemplate.Resource.Cpu != m.OBZone.Spec.OBServerTemplate.Resource.Cpu ||
			observer.Spec.OBServerTemplate.Resource.Memory != m.OBZone.Spec.OBServerTemplate.Resource.Memory {
			m.Logger.Info("scale up observer", "observer", observer.Name)
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				observer.Spec.OBServerTemplate.Resource.Cpu = m.OBZone.Spec.OBServerTemplate.Resource.Cpu
				observer.Spec.OBServerTemplate.Resource.Memory = m.OBZone.Spec.OBServerTemplate.Resource.Memory
				return m.Client.Update(m.Ctx, &observer)
			})
			if err != nil {
				return errors.Wrapf(err, "Scale up observer %s failed", observer.Name)
			}
		}
	}
	return nil
}

func (m *OBZoneManager) ResizePVC() tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	for _, observer := range observerList.Items {
		if observer.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 ||
			observer.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 ||
			observer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0 {

			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				m.Logger.Info("Expand pvc of observer", "observer", observer.Name)
				observer.Spec.OBServerTemplate.Storage.DataStorage.Size = m.OBZone.Spec.OBServerTemplate.Storage.DataStorage.Size
				observer.Spec.OBServerTemplate.Storage.LogStorage.Size = m.OBZone.Spec.OBServerTemplate.Storage.LogStorage.Size
				observer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size = m.OBZone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size
				return m.Client.Update(m.Ctx, &observer)
			})
			if err != nil {
				return errors.Wrapf(err, "Expand observer %s failed", observer.Name)
			}
		}
	}
	return nil
}

func (m *OBZoneManager) MountBackupVolume() tasktypes.TaskError {
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
