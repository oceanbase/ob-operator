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

package obcluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func (m *OBClusterManager) checkIfStorageSizeExpand(obzone *v1alpha1.OBZone) bool {
	newStorage := m.OBCluster.Spec.OBServerTemplate.Storage
	oldStorage := obzone.Spec.OBServerTemplate.Storage
	return oldStorage.DataStorage.Size.Cmp(newStorage.DataStorage.Size) < 0 ||
		oldStorage.LogStorage.Size.Cmp(newStorage.LogStorage.Size) < 0 ||
		oldStorage.RedoLogStorage.Size.Cmp(newStorage.RedoLogStorage.Size) < 0
}

func (m *OBClusterManager) checkIfStorageClassChange(obzone *v1alpha1.OBZone) bool {
	newStorage := m.OBCluster.Spec.OBServerTemplate.Storage
	oldStorage := obzone.Spec.OBServerTemplate.Storage
	return oldStorage.DataStorage.StorageClass != newStorage.DataStorage.StorageClass ||
		oldStorage.LogStorage.StorageClass != newStorage.LogStorage.StorageClass ||
		oldStorage.RedoLogStorage.StorageClass != newStorage.RedoLogStorage.StorageClass
}

func (m *OBClusterManager) checkIfCalcResourceChange(obzone *v1alpha1.OBZone) bool {
	return obzone.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBCluster.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		obzone.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBCluster.Spec.OBServerTemplate.Resource.Memory) != 0
}

func (m *OBClusterManager) checkIfBackupVolumeMutated(obzone *v1alpha1.OBZone) bool {
	return (obzone.Spec.BackupVolume == nil) != (m.OBCluster.Spec.BackupVolume == nil)
}

func (m *OBClusterManager) checkIfMonitorMutated(obzone *v1alpha1.OBZone) bool {
	return (obzone.Spec.MonitorTemplate == nil) != (m.OBCluster.Spec.MonitorTemplate == nil)
}

func (m *OBClusterManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		obcluster.Status = *m.OBCluster.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, obcluster)
	})
}

func (m *OBClusterManager) listOBZones() (*v1alpha1.OBZoneList, error) {
	// this label always exists
	obzoneList := &v1alpha1.OBZoneList{}
	err := m.Client.List(m.Ctx, obzoneList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: m.OBCluster.Name,
	}, client.InNamespace(m.OBCluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "get obzone list")
	}
	return obzoneList, nil
}

func (m *OBClusterManager) listOBParameters() (*v1alpha1.OBParameterList, error) {
	// this label always exists
	obparameterList := &v1alpha1.OBParameterList{}
	err := m.Client.List(m.Ctx, obparameterList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: m.OBCluster.Name,
	}, client.InNamespace(m.OBCluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "get obzone list")
	}
	return obparameterList, nil
}

func (m *OBClusterManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBCluster.Namespace,
		Name:      m.OBCluster.Name,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBClusterManager) generateZoneName(zone string) string {
	return fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, zone)
}

func (m *OBClusterManager) generateParameterName(name string) string {
	return fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, strings.ReplaceAll(name, "_", "-"))
}

func (m *OBClusterManager) getZonesToDelete() ([]v1alpha1.OBZone, error) {
	deletedZones := make([]v1alpha1.OBZone, 0)
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "List obzone failed")
		return deletedZones, errors.Wrapf(err, "List obzone of obcluster %s failed", m.OBCluster.Name)
	}
	for _, obzone := range obzoneList.Items {
		reserve := false
		for _, zone := range m.OBCluster.Spec.Topology {
			if zone.Zone == obzone.Spec.Topology.Zone {
				reserve = true
				break
			}
		}
		if !reserve {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Need to delete obzone", "obzone", obzone.Name)
			deletedZones = append(deletedZones, obzone)
		}
	}
	return deletedZones, nil
}

func (m *OBClusterManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	return resourceutils.GetSysOperationClient(m.Client, m.Logger, m.OBCluster)
}

func (m *OBClusterManager) createUser(userName, secretName, privilege string) error {
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Begin to create user", "username", userName)
	password, err := resourceutils.ReadPassword(m.Client, m.OBCluster.Namespace, secretName)
	if err != nil {
		return errors.Wrapf(err, "Get password from secret %s failed", secretName)
	}
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	err = oceanbaseOperationManager.CreateUser(m.Ctx, userName)
	if err != nil {
		m.Logger.Error(err, "Create user")
		return errors.Wrapf(err, "Create user %s", userName)
	}
	err = oceanbaseOperationManager.SetUserPassword(m.Ctx, userName, password)
	if err != nil {
		m.Logger.Error(err, "Set user password")
		return errors.Wrapf(err, "Set password for user %s", userName)
	}
	object := "*.*"
	err = oceanbaseOperationManager.GrantPrivilege(m.Ctx, privilege, object, userName)
	if err != nil {
		m.Logger.Error(err, "Grant privilege")
		return errors.Wrapf(err, "Grant privilege for user %s", userName)
	}
	return nil
}

type obzoneChanger func(*v1alpha1.OBZone)

func (m *OBClusterManager) changeZonesWhenScaling(obzone *v1alpha1.OBZone) {
	obzone.Spec.OBServerTemplate.Resource.Cpu = m.OBCluster.Spec.OBServerTemplate.Resource.Cpu
	obzone.Spec.OBServerTemplate.Resource.Memory = m.OBCluster.Spec.OBServerTemplate.Resource.Memory
}

func (m *OBClusterManager) changeZonesWhenExpandingPVC(obzone *v1alpha1.OBZone) {
	obzone.Spec.OBServerTemplate.Storage.DataStorage.Size = m.OBCluster.Spec.OBServerTemplate.Storage.DataStorage.Size
	obzone.Spec.OBServerTemplate.Storage.LogStorage.Size = m.OBCluster.Spec.OBServerTemplate.Storage.LogStorage.Size
	obzone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size = m.OBCluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size
}

func (m *OBClusterManager) changeZonesWhenUpdatingOBServers(obzone *v1alpha1.OBZone) {
	obzone.Spec.OBServerTemplate = m.OBCluster.Spec.OBServerTemplate
}

func (m *OBClusterManager) changeZonesWhenModifyingServerTemplate(obzone *v1alpha1.OBZone) {
	obzone.Spec.BackupVolume = m.OBCluster.Spec.BackupVolume
	obzone.Spec.MonitorTemplate = m.OBCluster.Spec.MonitorTemplate
}

func (m *OBClusterManager) modifyOBZonesAndCheckStatus(changer obzoneChanger, status string, timeoutSeconds int) tasktypes.TaskFunc {
	return func() tasktypes.TaskError {
		obzoneList, err := m.listOBZones()
		if err != nil {
			return errors.Wrap(err, "list obzones")
		}
		for _, obzone := range obzoneList.Items {
			changer(&obzone)
			err = m.Client.Update(m.Ctx, &obzone)
			if err != nil {
				return errors.Wrap(err, "update obzone")
			}
		}

		// check status of obzones
		matched := true
	outer:
		for i := 0; i < timeoutSeconds; i++ {
			time.Sleep(time.Second)
			obzoneList, err = m.listOBZones()
			if err != nil {
				return errors.Wrap(err, "list obzones")
			}
			for _, obzone := range obzoneList.Items {
				if obzone.Status.Status != status {
					matched = false
					continue outer
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return errors.New("failed to wait for status of obzone to be " + status)
		}
		return nil
	}
}

func (m *OBClusterManager) rollingUpdateZones(changer obzoneChanger, workingStatus, targetStatus string, timeoutSeconds int) tasktypes.TaskFunc {
	return func() tasktypes.TaskError {
		tk := time.NewTicker(time.Duration(timeoutSeconds*2) * time.Second)
		defer tk.Stop()
		obzoneList, err := m.listOBZones()
		if err != nil {
			return errors.Wrap(err, "list obzones")
		}
		for _, obzone := range obzoneList.Items {
			m.Recorder.Event(m.OBCluster, "Normal", "RollingUpdateOBZone", "Rolling update OBZone "+obzone.Name)
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				targetZone := v1alpha1.OBZone{}
				m.Client.Get(m.Ctx, types.NamespacedName{
					Namespace: m.OBCluster.Namespace,
					Name:      obzone.Name,
				}, &targetZone)
				changer(&targetZone)
				return m.Client.Update(m.Ctx, &targetZone)
			})
			if err != nil {
				return errors.Wrap(err, "update obzone")
			}
			for i := 0; i < timeoutSeconds; i++ {
				select {
				case <-tk.C:
					return errors.New("task timeout")
				default:
				}
				time.Sleep(time.Second)
				updatedOBZone := &v1alpha1.OBZone{}
				err := m.Client.Get(m.Ctx, types.NamespacedName{
					Namespace: obzone.Namespace,
					Name:      obzone.Name,
				}, updatedOBZone)
				if err != nil {
					return errors.Wrap(err, "get obzone")
				}
				if updatedOBZone.Status.Status == workingStatus {
					break
				}
			}
			for i := 0; i < timeoutSeconds; i++ {
				select {
				case <-tk.C:
					return errors.New("task timeout")
				default:
				}
				time.Sleep(time.Second)
				updatedOBZone := &v1alpha1.OBZone{}
				err := m.Client.Get(m.Ctx, types.NamespacedName{
					Namespace: obzone.Namespace,
					Name:      obzone.Name,
				}, updatedOBZone)
				if err != nil {
					return errors.Wrap(err, "get obzone")
				}
				if updatedOBZone.Status.Status == targetStatus {
					break
				}
			}
		}
		return nil
	}
}

func (m *OBClusterManager) generateWaitOBZoneStatusFunc(status string, timeoutSeconds int) tasktypes.TaskFunc {
	f := func() tasktypes.TaskError {
		for i := 1; i < timeoutSeconds; i++ {
			obcluster, err := m.getOBCluster()
			if err != nil {
				return errors.Wrap(err, "get obcluster failed")
			}
			allMatched := true
			for _, obzoneStatus := range obcluster.Status.OBZoneStatus {
				if obzoneStatus.Status != status {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Zone status still not matched", "zone", obzoneStatus.Zone, "status", status)
					allMatched = false
					break
				}
			}
			if allMatched {
				return nil
			}
			time.Sleep(time.Second)
		}
		return errors.New("Zone status still not matched when timeout")
	}
	return f
}

func (m *OBClusterManager) CreateOBParameter(parameter *apitypes.Parameter) error {
	m.Logger.Info("Create ob parameters")
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBCluster.APIVersion,
		Kind:       m.OBCluster.Kind,
		Name:       m.OBCluster.Name,
		UID:        m.OBCluster.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	labels := make(map[string]string)
	labels[oceanbaseconst.LabelRefUID] = string(m.OBCluster.GetUID())
	labels[oceanbaseconst.LabelRefOBCluster] = m.OBCluster.Name
	parameterName := m.generateParameterName(parameter.Name)
	obparameter := &v1alpha1.OBParameter{
		ObjectMeta: metav1.ObjectMeta{
			Name:            parameterName,
			Namespace:       m.OBCluster.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          labels,
		},
		Spec: v1alpha1.OBParameterSpec{
			ClusterName: m.OBCluster.Spec.ClusterName,
			ClusterId:   m.OBCluster.Spec.ClusterId,
			Parameter:   parameter,
		},
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Create obparameter", "parameter", parameterName)
	err := m.Client.Create(m.Ctx, obparameter)
	if err != nil {
		m.Logger.Error(err, "create obparameter failed")
		return errors.Wrap(err, "create obparameter")
	}
	return nil
}

func (m *OBClusterManager) UpdateOBParameter(parameter *apitypes.Parameter) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obparameter := &v1alpha1.OBParameter{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBCluster.Namespace,
			Name:      m.generateParameterName(parameter.Name),
		}, obparameter)
		if err != nil {
			return errors.Wrap(err, "Get obparameter")
		}
		obparameter.Spec.Parameter.Value = parameter.Value
		err = m.Client.Update(m.Ctx, obparameter)
		if err != nil {
			return errors.Wrap(err, "Update obparameter")
		}
		return nil
	})
}

func (m *OBClusterManager) DeleteOBParameter(parameter *apitypes.Parameter) error {
	obparameter := &v1alpha1.OBParameter{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBCluster.Namespace,
		Name:      m.generateParameterName(parameter.Name),
	}, obparameter)
	if err != nil {
		return errors.Wrap(err, "Get obparameter")
	}
	obparameter.Spec.Parameter.Value = parameter.Value
	err = m.Client.Delete(m.Ctx, obparameter)
	if err != nil {
		return errors.Wrap(err, "Delete obparameter")
	}
	return nil
}

func (m *OBClusterManager) WaitOBZoneUpgradeFinished(zoneName string) error {
	check := func() (bool, error) {
		zones, err := m.listOBZones()
		if err != nil {
			return false, errors.Wrap(err, "Failed to get obzone list")
		}
		for _, zone := range zones.Items {
			if zone.Name != zoneName {
				continue
			}
			m.Logger.Info("Check obzone upgrade status", "obzone", zoneName)
			if zone.Status.Status == zonestatus.Running && zone.Status.Image == m.OBCluster.Spec.OBServerTemplate.Image {
				m.Logger.Info("OBZone upgrade finished", "obzone", zoneName)
				return true, nil
			}
		}
		return false, nil
	}
	err := resourceutils.CheckJobWithTimeout(check, time.Second*time.Duration(obcfg.GetConfig().Time.WaitForJobTimeoutSeconds))
	if err != nil {
		return errors.Wrap(err, "Timeout to wait obzone upgrade finished")
	}
	return nil
}
