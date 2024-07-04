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
	pq "github.com/ugurcsen/gods-generic/queues/priorityqueue"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

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
	currentReplica := 0
	for _, observerStatus := range m.OBZone.Status.OBServerStatus {
		if observerStatus.Status != serverstatus.Unrecoverable {
			currentReplica++
		}
	}
	for i := currentReplica; i < m.OBZone.Spec.Topology.Replica; i++ {
		serverName := m.generateServerName()
		_, err := m.createOneOBServer(serverName)
		if err != nil {
			return errors.Wrapf(err, "Create observer %s", serverName)
		}
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
	queue := pq.NewWith(resourceutils.CompOBServerDeletionPriority)
	for _, observer := range observerList.Items {
		queue.Enqueue(&observer)
	}
	for !queue.Empty() {
		observer, _ := queue.Dequeue()
		if observer.Status.Status == serverstatus.Deleting {
			continue
		}
		if observer.Status.Status == serverstatus.Unrecoverable || observerCount >= m.OBZone.Spec.Topology.Replica {
			m.Logger.Info("Delete observer", "observer", observer)
			err = m.Client.Delete(m.Ctx, observer)
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

func ScaleOBServersVertically(m *OBZoneManager) tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	zoneRes := m.OBZone.Spec.OBServerTemplate.Resource
	for _, observer := range observerList.Items {
		serverRes := observer.Spec.OBServerTemplate.Resource
		if serverRes.Cpu != zoneRes.Cpu || serverRes.Memory != zoneRes.Memory {
			m.Logger.Info("Scale observer vertically", "observer", observer.Name)
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				serverRes.Cpu = zoneRes.Cpu
				serverRes.Memory = zoneRes.Memory
				return m.Client.Update(m.Ctx, &observer)
			})
			if err != nil {
				return errors.Wrapf(err, "Scale observer %s vertically failed", observer.Name)
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

func ModifyPodTemplate(m *OBZoneManager) tasktypes.TaskError {
	observerList, err := m.listOBServers()
	if err != nil {
		return err
	}
	for _, observer := range observerList.Items {
		observerCopied := observer.DeepCopy()
		if observer.Spec.BackupVolume == nil && m.OBZone.Spec.BackupVolume != nil {
			m.Logger.Info("Add backup volume", "observer", observer.Name)
			observer.Spec.BackupVolume = m.OBZone.Spec.BackupVolume
			err = m.Client.Patch(m.Ctx, &observer, client.MergeFrom(observerCopied))
			if err != nil {
				return errors.Wrapf(err, "Add backup volume %s failed", observer.Name)
			}
		}
		if observer.Spec.MonitorTemplate == nil && m.OBZone.Spec.MonitorTemplate != nil {
			m.Logger.Info("Add monitor template", "observer", observer.Name)
			observer.Spec.MonitorTemplate = m.OBZone.Spec.MonitorTemplate
			err = m.Client.Patch(m.Ctx, &observer, client.MergeFrom(observerCopied))
			if err != nil {
				return errors.Wrapf(err, "Add monitor template %s failed", observer.Name)
			}
		}
		if observer.Spec.BackupVolume != nil && m.OBZone.Spec.BackupVolume == nil {
			m.Logger.Info("Remove backup volume", "observer", observer.Name)
			observer.Spec.BackupVolume = nil
			err = m.Client.Patch(m.Ctx, &observer, client.MergeFrom(observerCopied))
			if err != nil {
				return errors.Wrapf(err, "Remove backup volume %s failed", observer.Name)
			}
		}
		if observer.Spec.MonitorTemplate != nil && m.OBZone.Spec.MonitorTemplate == nil {
			m.Logger.Info("Remove monitor template", "observer", observer.Name)
			observer.Spec.MonitorTemplate = nil
			err = m.Client.Patch(m.Ctx, &observer, client.MergeFrom(observerCopied))
			if err != nil {
				return errors.Wrapf(err, "Remove monitor template %s failed", observer.Name)
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
	return m.generateWaitOBServerStatusFunc(serverstatus.ScaleVertically, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerExpandingPVC(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.ExpandPVC, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerTemplateModifying(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.ModifyingPodTemplate, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func RollingReplaceOBServers(m *OBZoneManager) tasktypes.TaskError {
	servers, err := m.listOBServers()
	if err != nil {
		return errors.Wrap(err, "List observers")
	}
	for _, server := range servers.Items {
		newServerName := m.generateServerName()
		newServer, err := m.createOneOBServer(newServerName)
		if err != nil {
			return errors.Wrap(err, "Create new observer to replace old one")
		}
		for i := 0; i < obcfg.GetConfig().Time.DefaultStateWaitTimeout; i++ {
			time.Sleep(time.Second)
			err = m.Client.Get(m.Ctx, m.generateNamespacedName(newServerName), newServer)
			if err != nil {
				return errors.Wrap(err, "Get new observer")
			}
			if newServer.Status.Status == serverstatus.Running {
				break
			}
		}
		if newServer.Status.Status != serverstatus.Running {
			return errors.New("Wait for new observer get running status, timeout")
		}
		err = m.Client.Delete(m.Ctx, &server)
		if err != nil {
			return errors.Wrap(err, "Delete old observer")
		}
		for i := 0; i < obcfg.GetConfig().Time.DefaultStateWaitTimeout; i++ {
			time.Sleep(time.Second)
			oldServer := &v1alpha1.OBServer{}
			err = m.Client.Get(m.Ctx, m.generateNamespacedName(server.Name), oldServer)
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					break
				}
				return errors.Wrap(err, "Get old observer")
			}
		}
	}
	return nil
}
