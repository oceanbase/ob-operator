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

package resource

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/pkg/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

func (m *OBZoneManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return GetSysOperationClient(m.Client, m.Logger, obcluster)
}

func (m *OBZoneManager) generateServerName() string {
	parts := strings.Split(uuid.New().String(), "-")
	suffix := parts[len(parts)-1]
	return fmt.Sprintf("%s-%d-%s-%s", m.OBZone.Spec.ClusterName, m.OBZone.Spec.ClusterId, m.OBZone.Spec.Topology.Zone, suffix)
}

func (m *OBZoneManager) AddZone() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	return oceanbaseOperationManager.AddZone(m.OBZone.Spec.Topology.Zone)
}

func (m *OBZoneManager) StartOBZone() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	return oceanbaseOperationManager.StartZone(m.OBZone.Spec.Topology.Zone)
}

func (m *OBZoneManager) generateWaitOBServerStatusFunc(status string, timeoutSeconds int) func() error {
	f := func() error {
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

func (m *OBZoneManager) CreateOBServer() error {
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
	sepVolumeAnnoVal, sepVolumeAnnoExist := GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsIndependentPVCLifecycle)
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
			},
		}
		if sepVolumeAnnoExist {
			observer.ObjectMeta.Annotations = make(map[string]string)
			observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsIndependentPVCLifecycle] = sepVolumeAnnoVal
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

func (m *OBZoneManager) DeleteOBServer() error {
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("delete observers")
	observerList, err := m.listOBServers()
	if err != nil {
		m.Logger.Error(err, "List observers failed")
		return errors.Wrapf(err, "List observrers of obzone %s", m.OBZone.Name)
	}
	observerCount := 0
	for _, observer := range observerList.Items {
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
func (m *OBZoneManager) DeleteAllOBServer() error {
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

func (m *OBZoneManager) WaitReplicaMatch() error {
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

func (m *OBZoneManager) WaitOBServerDeleted() error {
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

func (m *OBZoneManager) StopOBZone() error {
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

func (m *OBZoneManager) OBClusterHealthCheck() error {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	_ = ExecuteUpgradeScript(m.Client, m.Logger, obcluster, oceanbaseconst.UpgradeHealthCheckerScriptPath, "")
	return nil
}

func (m *OBZoneManager) OBZoneHealthCheck() error {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	zoneOpt := fmt.Sprintf("-z '%s'", m.OBZone.Spec.Topology.Zone)
	_ = ExecuteUpgradeScript(m.Client, m.Logger, obcluster, oceanbaseconst.UpgradeHealthCheckerScriptPath, zoneOpt)
	return nil
}

func (m *OBZoneManager) UpgradeOBServer() error {
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

func (m *OBZoneManager) WaitOBServerUpgraded() error {
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

func (m *OBZoneManager) DeleteOBZoneInCluster() error {
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
