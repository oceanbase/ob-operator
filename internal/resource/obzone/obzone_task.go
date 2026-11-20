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
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	pq "github.com/ugurcsen/gods-generic/queues/priorityqueue"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
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

func AddSharedStorageDest(m *OBZoneManager) tasktypes.TaskError {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster")
	}
	if obcluster.Spec.DeploymentMode != oceanbaseconst.DeploymentModeSharedStorage {
		return nil
	}
	if obcluster.Spec.SharedStorageInfo == nil {
		return errors.New("sharedStorageInfo is required for shared_storage mode")
	}

	secret, err := resourceutils.GetSecret(m.Client, obcluster.Namespace, obcluster.Spec.SharedStorageInfo.SecretRef.Name)
	if err != nil {
		return errors.Wrap(err, "Get shared storage secret")
	}
	accessID, ok := secret.Data["access_id"]
	if !ok || len(accessID) == 0 {
		return errors.New("access_id is required in shared storage secret")
	}
	accessKey, ok := secret.Data["access_key"]
	if !ok || len(accessKey) == 0 {
		return errors.New("access_key is required in shared storage secret")
	}

	// Use the same runtime endpoint as bootstrap, without changing the CR spec.
	bucketURL, err := resourceutils.ResolveStorageServiceURL(m.Ctx, m.Client, obcluster, obcluster.Spec.SharedStorageInfo.BucketURL)
	if err != nil {
		return errors.Wrap(err, "Resolve shared storage endpoint")
	}

	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	accessInfo := fmt.Sprintf("access_id=%s&access_key=%s", accessID, accessKey)
	attribute, err := sharedStorageAttributeFromSpec(obcluster.Spec.SharedStorageInfo)
	if err != nil {
		return err
	}
	return oceanbaseOperationManager.AddSharedStorageDest(
		m.Ctx,
		bucketURL,
		accessInfo,
		attribute,
		m.OBZone.Spec.Topology.Zone,
	)
}

func buildSharedStorageAttribute(maxIOPS, maxBandwidth string) string {
	maxIOPS, maxBandwidth = strings.TrimSpace(maxIOPS), strings.TrimSpace(maxBandwidth)
	if maxIOPS == "" {
		maxIOPS = "0"
	}
	if maxBandwidth == "" {
		maxBandwidth = "0B"
	} else if _, err := strconv.ParseUint(maxBandwidth, 10, 64); err == nil {
		maxBandwidth += "B"
	}
	return "max_iops=" + maxIOPS + "&max_bandwidth=" + maxBandwidth
}

func sharedStorageAttributeFromSpec(storage *apitypes.SharedStorageSpec) (string, error) {
	// Preserve limits embedded in bucketURL, as accepted by BOOTSTRAP. Explicit
	// fields take precedence. Do not log the URL: it may contain credentials.
	parsed, err := url.Parse(storage.BucketURL)
	if err != nil {
		return "", errors.New("invalid shared storage bucket URL")
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", errors.New("invalid shared storage bucket URL query")
	}
	maxIOPS, maxBandwidth := storage.MaxIOPS, storage.MaxBandwidth
	if strings.TrimSpace(maxIOPS) == "" {
		maxIOPS = query.Get("max_iops")
	}
	if strings.TrimSpace(maxBandwidth) == "" {
		maxBandwidth = query.Get("max_bandwidth")
	}
	return buildSharedStorageAttribute(maxIOPS, maxBandwidth), nil
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
		_, err := m.createOneOBServer(serverName, "")
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
	cluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster before deleting zone")
	}
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "OBZone %s get oceanbase operation manager", m.OBZone.Name)
	}
	if cluster.Spec.DeploymentMode == oceanbaseconst.DeploymentModeSharedStorage {
		if err := operationManager.DropZoneSharedStorage(m.Ctx, m.OBZone.Spec.Topology.Zone); err != nil {
			return errors.Wrap(err, "Drop shared storage before deleting zone")
		}
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
			(serverStorage.RedoLogStorage != nil && zoneStorage.RedoLogStorage != nil &&
				serverStorage.RedoLogStorage.Size.Cmp(zoneStorage.RedoLogStorage.Size) < 0) {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				m.Logger.Info("Expand pvc of observer", "observer", observer.Name)
				serverStorage.DataStorage.Size = zoneStorage.DataStorage.Size
				serverStorage.LogStorage.Size = zoneStorage.LogStorage.Size
				if serverStorage.RedoLogStorage != nil && zoneStorage.RedoLogStorage != nil {
					serverStorage.RedoLogStorage.Size = zoneStorage.RedoLogStorage.Size
				}
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
		if !m.ownsRollingServer(&observer) || observer.DeletionTimestamp != nil {
			return errors.Errorf("observer %s is deleting or has a different owner", observer.Name)
		}
		backupChanged := m.checkIfBackupVolumeMutated(&observer)
		monitorChanged := m.checkIfMonitorMutated(&observer)
		if !backupChanged && !monitorChanged {
			continue
		}
		original := observer.DeepCopy()
		if backupChanged {
			observer.Spec.BackupVolume = m.OBZone.Spec.BackupVolume
		}
		if monitorChanged {
			observer.Spec.MonitorTemplate = m.OBZone.Spec.Topology.EffectiveMonitor(m.OBZone.Spec.MonitorTemplate)
			if observer.Annotations == nil {
				observer.Annotations = make(map[string]string)
			}
			// Record the edit together with the spec. This also distinguishes an
			// intentional CPU limit removal from a limit injected by admission.
			observer.Annotations[oceanbaseconst.AnnotationsMonitorRevision] = original.ResourceVersion
		}
		if err := m.Client.Patch(m.Ctx, &observer, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{})); err != nil {
			return errors.Wrapf(err, "update pod template for observer %s", observer.Name)
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
	return m.waitForBootstrapReplicas(serverstatus.BootstrapReady)
}

func WaitOBServerRunning(m *OBZoneManager) tasktypes.TaskError {
	if m.OBZone.Status.Status == zonestatus.BootstrapReady {
		return m.waitForBootstrapReplicas(serverstatus.Running)
	}
	return m.generateWaitOBServerStatusFunc(serverstatus.Running, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerScalingUp(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.ScaleVertically, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerExpandingPVC(m *OBZoneManager) tasktypes.TaskError {
	return m.generateWaitOBServerStatusFunc(serverstatus.ExpandPVC, obcfg.GetConfig().Time.DefaultStateWaitTimeout)()
}

func WaitForOBServerTemplateModifying(m *OBZoneManager) tasktypes.TaskError {
	timeout := time.Duration(obcfg.GetConfig().Time.DefaultStateWaitTimeout) * time.Second
	return wait.PollUntilContextTimeout(m.Ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		return resourceutils.ZonePodTemplatesReady(ctx, m.Client, m.OBZone)
	})
}

func RollingReplaceOBServers(m *OBZoneManager) tasktypes.TaskError {
	servers, err := m.listOBServers()
	if err != nil {
		return errors.Wrap(err, "List observers")
	}
	// A finalizer may still be draining a server from the preceding attempt.
	for _, server := range servers.Items {
		if !m.ownsRollingServer(&server) {
			return errors.Errorf("observer %s does not belong to zone UID %s", server.Name, m.OBZone.UID)
		}
		if server.DeletionTimestamp != nil {
			if err := m.waitForRollingServerDeletion(server.Name); err != nil {
				return err
			}
		}
	}
	prioritizePendingReplacements(servers.Items)
	for _, server := range servers.Items {
		if server.DeletionTimestamp != nil || !m.checkIfStorageClassChanged(&server) {
			continue
		}
		target, err := m.getRollingReplacement(&server)
		if err != nil {
			return err
		}
		if target == nil {
			continue
		}
		if err := m.waitForRollingReplacement(target, server.UID); err != nil {
			return err
		}
		err = m.Client.Delete(m.Ctx, &server, client.Preconditions{UID: &server.UID})
		if client.IgnoreNotFound(err) != nil {
			return errors.Wrapf(err, "delete replaced observer %s", server.Name)
		}
		if err := m.waitForRollingServerDeletion(server.Name); err != nil {
			return err
		}
	}
	return nil
}

func (m *OBZoneManager) waitForRollingServerDeletion(name string) error {
	timeout := time.Duration(obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds) * time.Second
	err := wait.PollUntilContextTimeout(m.Ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		err := m.Client.Get(ctx, m.generateNamespacedName(name), &v1alpha1.OBServer{})
		if kubeerrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
	return errors.Wrapf(err, "wait for deletion of observer %s", name)
}
