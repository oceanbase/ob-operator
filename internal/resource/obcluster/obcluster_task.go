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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
	obutil "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/util"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func (m *OBClusterManager) WaitOBZoneTopologyMatch() tasktypes.TaskError {
	// TODO
	return nil
}

func (m *OBClusterManager) WaitOBZoneDeleted() tasktypes.TaskError {
	waitSuccess := false
	for i := 1; i < oceanbaseconst.ServerDeleteTimeoutSeconds; i++ {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return errors.Wrap(err, "get obcluster failed")
		}
		zoneDeleted := true
		for _, zoneStatus := range obcluster.Status.OBZoneStatus {
			found := false
			for _, zone := range m.OBCluster.Spec.Topology {
				if zoneStatus.Zone == m.generateZoneName(zone.Zone) {
					found = true
					break
				}
			}
			if !found {
				m.Logger.Info("OBZone not in spec, still not deleted", "zone", zoneStatus.Zone)
				zoneDeleted = false
				break
			}
		}
		if zoneDeleted {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("All zone deleted")
			waitSuccess = true
			break
		}
		time.Sleep(time.Second * 1)
	}
	if waitSuccess {
		return nil
	}
	return errors.Errorf("OBCluster %s zone still not deleted when timeout", m.OBCluster.Name)
}

func (m *OBClusterManager) generateWaitOBZoneStatusFunc(status string, timeoutSeconds int) func() tasktypes.TaskError {
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

func (m *OBClusterManager) ModifyOBZoneReplica() tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obzoneList, err := m.listOBZones()
		if err != nil {
			m.Logger.Error(err, "List obzone failed")
			return errors.Wrapf(err, "List obzone of obcluster %s failed", m.OBCluster.Name)
		}
		for _, zone := range m.OBCluster.Spec.Topology {
			for _, obzone := range obzoneList.Items {
				if zone.Zone == obzone.Spec.Topology.Zone && zone.Replica != obzone.Spec.Topology.Replica {
					m.Logger.Info("Modify obzone replica", "obzone", zone.Zone)
					obzone.Spec.Topology.Replica = zone.Replica
					err = m.Client.Update(m.Ctx, &obzone)
					if err != nil {
						return errors.Wrapf(err, "Modify obzone %s replica failed", zone.Zone)
					}
				}
			}
		}
		return nil
	})
}

func (m *OBClusterManager) DeleteOBZone() tasktypes.TaskError {
	zonesToDelete, err := m.getZonesToDelete()
	if err != nil {
		return errors.Wrap(err, "Failed to get obzones to delete")
	}
	for _, zone := range zonesToDelete {
		err = m.Client.Delete(m.Ctx, &zone)
		if err != nil {
			return errors.Wrapf(err, "Delete obzone %s", zone.Name)
		}
		m.Recorder.Event(m.OBCluster, "DeleteOBZone", "", fmt.Sprintf("Delete obzone %s successfully", zone.Name))
	}
	return nil
}

func (m *OBClusterManager) CreateOBZone() tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Create obzones")
	blockOwnerDeletion := true
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion:         m.OBCluster.APIVersion,
		Kind:               m.OBCluster.Kind,
		Name:               m.OBCluster.Name,
		UID:                m.OBCluster.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	independentVolumeAnnoVal, independentVolumeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsIndependentPVCLifecycle)
	singlePVCAnnoVal, singlePVCAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsSinglePVC)
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)
	migrateAnnoVal, migrateAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsSourceClusterAddress)
	for _, zone := range m.OBCluster.Spec.Topology {
		zoneName := m.generateZoneName(zone.Zone)
		zoneExists := false
		for _, zoneStatus := range m.OBCluster.Status.OBZoneStatus {
			if zoneName == zoneStatus.Zone {
				zoneExists = true
				break
			}
		}
		if zoneExists {
			m.Logger.Info("Zone already exists", "zone", zoneName)
			continue
		}
		labels := make(map[string]string)
		labels[oceanbaseconst.LabelRefUID] = string(m.OBCluster.GetUID())
		labels[oceanbaseconst.LabelRefOBCluster] = m.OBCluster.Name
		finalizerName := "finalizers.oceanbase.com.deleteobzone"
		finalizers := []string{finalizerName}
		obzone := &v1alpha1.OBZone{
			ObjectMeta: metav1.ObjectMeta{
				Name:            zoneName,
				Namespace:       m.OBCluster.Namespace,
				OwnerReferences: ownerReferenceList,
				Labels:          labels,
				Finalizers:      finalizers,
			},
			Spec: v1alpha1.OBZoneSpec{
				ClusterName:      m.OBCluster.Spec.ClusterName,
				ClusterId:        m.OBCluster.Spec.ClusterId,
				OBServerTemplate: m.OBCluster.Spec.OBServerTemplate,
				MonitorTemplate:  m.OBCluster.Spec.MonitorTemplate,
				BackupVolume:     m.OBCluster.Spec.BackupVolume,
				Topology:         zone,
				ServiceAccount:   m.OBCluster.Spec.ServiceAccount,
			},
		}
		obzone.ObjectMeta.Annotations = make(map[string]string)
		if independentVolumeAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsIndependentPVCLifecycle] = independentVolumeAnnoVal
		}
		if singlePVCAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSinglePVC] = singlePVCAnnoVal
		}
		if modeAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsMode] = modeAnnoVal
		}
		if migrateAnnoExist {
			obzone.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSourceClusterAddress] = migrateAnnoVal
		}
		m.Logger.Info("Create obzone", "zone", zoneName)
		err := m.Client.Create(m.Ctx, obzone)
		if err != nil {
			m.Logger.Error(err, "create obzone failed", "zone", zone.Zone)
			return errors.Wrap(err, "create obzone")
		}
		m.Recorder.Event(m.OBCluster, "CreateOBZone", "", fmt.Sprintf("Create obzone %s successfully", zoneName))
	}
	return nil
}

func (m *OBClusterManager) Bootstrap() tasktypes.TaskError {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones failed")
		return errors.Wrap(err, "list obzones")
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Successfully get obzone list", "obzone list", obzoneList)
	if len(obzoneList.Items) == 0 {
		return errors.Wrap(err, "no obzone belongs to this cluster")
	}
	var manager *operation.OceanbaseOperationManager
	for i := 0; i < oceanbaseconst.GetConnectionMaxRetries; i++ {
		manager, err = m.getOceanbaseOperationManager()
		if err != nil || manager == nil {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Get oceanbase operation manager failed")
			time.Sleep(time.Second * oceanbaseconst.CheckConnectionInterval)
		} else {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Successfully got oceanbase operation manager")
			break
		}
	}
	if err != nil {
		m.Logger.Error(err, "get oceanbase operation manager failed")
		return errors.Wrap(err, "get oceanbase operation manager")
	}

	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)

	bootstrapServers := make([]model.BootstrapServerInfo, 0, len(m.OBCluster.Spec.Topology))
	if modeAnnoExist && modeAnnoVal == oceanbaseconst.ModeStandalone {
		m.Logger.Info("Bootstrap as standalone mode")
		bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
			Zone: m.OBCluster.Spec.Topology[0].Zone,
			Server: &model.ServerInfo{
				Ip:   "127.0.0.1",
				Port: oceanbaseconst.RpcPort,
			},
		})
	} else {
		connectAddress := manager.Connector.DataSource().GetAddress()
		for _, zone := range obzoneList.Items {
			serverIp := zone.Status.OBServerStatus[0].GetConnectAddr()
			// Notes: If the addr of the db connector is in this obzone, use it as the bootstrap server instead of the first one
			for _, serverInfo := range zone.Status.OBServerStatus {
				if serverInfo.Server == connectAddress || serverInfo.ServiceIP == connectAddress {
					serverIp = serverInfo.GetConnectAddr()
				}
			}
			serverInfo := &model.ServerInfo{
				Ip:   serverIp,
				Port: oceanbaseconst.RpcPort,
			}
			bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
				Zone:   zone.Spec.Topology.Zone,
				Server: serverInfo,
			})
		}
	}

	err = manager.Bootstrap(bootstrapServers)
	if err != nil {
		m.Logger.Error(err, "bootstrap failed")
	} else {
		m.Recorder.Event(m.OBCluster, "Bootstrap", "", "Bootstrap successfully")
	}
	return err
}

// Use Or for compatibility
func (m *OBClusterManager) CreateUsers() tasktypes.TaskError {
	err := m.createUser(oceanbaseconst.RootUser, m.OBCluster.Spec.UserSecrets.Root, oceanbaseconst.AllPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	err = m.createUser(oceanbaseconst.OperatorUser, m.OBCluster.Spec.UserSecrets.Operator, oceanbaseconst.AllPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create operator user")
	}
	err = m.createUser(obagentconst.MonitorUser, m.OBCluster.Spec.UserSecrets.Monitor, oceanbaseconst.SelectPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create monitor user")
	}
	err = m.createUser(oceanbaseconst.ProxyUser, m.OBCluster.Spec.UserSecrets.ProxyRO, oceanbaseconst.SelectPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create proxyro user")
	}
	return nil
}

func (m *OBClusterManager) MaintainOBParameter() tasktypes.TaskError {
	parameterMap := make(map[string]apitypes.Parameter)
	for _, parameter := range m.OBCluster.Status.Parameters {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Build parameter map", "parameter", parameter.Name)
		parameterMap[parameter.Name] = parameter
	}
	for _, parameter := range m.OBCluster.Spec.Parameters {
		parameterStatus, parameterExists := parameterMap[parameter.Name]
		if !parameterExists {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Parameter not exists, need create", "param", parameter.Name)
			err := m.CreateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Crate obparameter failed", "param", parameter.Name)
			}
		} else if parameterStatus.Value != parameter.Value {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Parameter value not matched, need update", "param", parameter.Name)
			err := m.UpdateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Update obparameter failed", "param", parameter.Name)
			}
		}
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Remove parameter from map", "parameter", parameter.Name)
		delete(parameterMap, parameter.Name)
	}

	// delete parameters that not in spec definition
	for _, parameter := range parameterMap {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Delete parameter", "parameter", parameter.Name)
		err := m.DeleteOBParameter(&parameter)
		if err != nil {
			m.Logger.Error(err, "Failed to delete parameter")
		}
	}
	return nil
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

func (m *OBClusterManager) ValidateUpgradeInfo() tasktypes.TaskError {
	// Get current obcluster version
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	// version, err := oceanbaseOperationManager.GetVersion()
	version, err := oceanbaseOperationManager.GetVersion()
	if err != nil {
		return errors.Wrapf(err, "Failed to get version of obcluster %s", m.OBCluster.Name)
	}
	// Get target version and patch
	parts := strings.Split(uuid.New().String(), "-")
	suffix := parts[len(parts)-1]
	jobName := fmt.Sprintf("%s-%s", "oceanbase-upgrade", suffix)
	var backoffLimit int32
	var ttl int32 = 300
	container := corev1.Container{
		Name:    "ob-upgrade-validator",
		Image:   m.OBCluster.Spec.OBServerTemplate.Image,
		Command: []string{"bash", "-c", fmt.Sprintf("/home/admin/oceanbase/bin/oceanbase-helper upgrade validate -s %s", version.String())},
	}
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: m.OBCluster.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{container},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
		},
	}

	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Create validate upgrade job", "job", jobName)
	err = m.Client.Create(m.Ctx, &job)
	if err != nil {
		return errors.Wrapf(err, "Failed to create validate job for obcluster %s", m.OBCluster.Name)
	}

	var jobObject *batchv1.Job
	for {
		time.Sleep(time.Second * oceanbaseconst.CheckJobInterval)
		jobObject, err = resourceutils.GetJob(m.Client, m.OBCluster.Namespace, jobName)
		if err != nil {
			m.Logger.Error(err, "Failed to get job")
		}
		if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Job is still running")
		} else {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Job finished")
			break
		}
	}

	if jobObject.Status.Succeeded == 1 {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Job succeeded")
	} else {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Job is failed", "job", jobName)
		return errors.Wrap(err, "Failed to run validate job")
	}
	return nil
}

func (m *OBClusterManager) UpgradeCheck() tasktypes.TaskError {
	return resourceutils.ExecuteUpgradeScript(m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradeCheckerScriptPath, "")
}

func (m *OBClusterManager) BackupEssentialParameters() tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	essentialParameters := make([]model.Parameter, 0)
	for _, parameter := range oceanbaseconst.UpgradeEssentialParameters {
		parameterValues, err := oceanbaseOperationManager.GetParameter(parameter, nil)
		if err != nil {
			return errors.Wrapf(err, "Failed to get parameter %s", parameter)
		}
		essentialParameters = append(essentialParameters, parameterValues...)
	}

	contextMap := make(map[string]string)
	jsonContent, err := json.Marshal(essentialParameters)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal essential parameters")
	}
	contextMap[oceanbaseconst.EssentialParametersKey] = string(jsonContent)
	contextObjectName := fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, oceanbaseconst.EssentialParametersKey)
	contextSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contextObjectName,
			Namespace: m.OBCluster.Namespace,
		},
		Type:       "Opaque",
		StringData: contextMap,
	}
	err = m.Client.Create(m.Ctx, contextSecret)
	if err != nil {
		return errors.Wrap(err, "Create context secret object")
	}
	return nil
}

func (m *OBClusterManager) BeginUpgrade() tasktypes.TaskError {
	return resourceutils.ExecuteUpgradeScript(m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradePreScriptPath, "")
}

// TODO: add timeout
func (m *OBClusterManager) WaitOBZoneUpgradeFinished(zoneName string) error {
	upgradeFinished := false
	for {
		zones, err := m.listOBZones()
		if err != nil {
			return errors.Wrap(err, "Failed to get obzone list")
		}
		for _, zone := range zones.Items {
			if zone.Name != zoneName {
				continue
			}
			m.Logger.Info("Check obzone upgrade status", "obzone", zoneName)
			if zone.Status.Status == zonestatus.Running && zone.Status.Image == m.OBCluster.Spec.OBServerTemplate.Image {
				upgradeFinished = true
				break
			}
		}
		if upgradeFinished {
			m.Logger.Info("OBZone upgrade finished", "obzone", zoneName)
			break
		}
		time.Sleep(time.Second * oceanbaseconst.CommonCheckInterval)
	}
	return nil
}

// TODO: add timeout
func (m *OBClusterManager) RollingUpgradeByZone() tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		zones, err := m.listOBZones()
		if err != nil {
			return errors.Wrap(err, "Failed to get obzone list")
		}
		for _, zone := range zones.Items {
			// update image and tag
			zone.Spec.OBServerTemplate.Image = m.OBCluster.Spec.OBServerTemplate.Image
			err = m.Client.Update(m.Ctx, &zone)
			if err != nil {
				return errors.Wrap(err, "Failed to update obzone image")
			}
			err = m.WaitOBZoneUpgradeFinished(zone.Name)
			if err != nil {
				return errors.Wrapf(err, "Wait obzone %s upgrade finish failed", zone.Name)
			}
		}
		return nil
	})
}

func (m *OBClusterManager) FinishUpgrade() tasktypes.TaskError {
	return resourceutils.ExecuteUpgradeScript(m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradePostScriptPath, "")
}

func (m *OBClusterManager) ModifySysTenantReplica() tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	desiredZones := make([]string, 0)
	for _, obzone := range m.OBCluster.Spec.Topology {
		desiredZones = append(desiredZones, obzone.Zone)
	}
	// add zone to pool zone list
	sysPool, err := oceanbaseOperationManager.GetPoolByName(oceanbaseconst.SysTenantPool)
	if err != nil {
		return errors.Wrap(err, "Failed to get sys pool info")
	}
	zoneList := strings.Split(sysPool.ZoneList, ";")
	for _, zone := range desiredZones {
		found := false
		for _, z := range zoneList {
			if zone == z {
				found = true
				break
			}
		}
		if !found {
			zoneList = append(zoneList, zone)
		}
	}
	m.Logger.Info("Modify sys pool's zone list when add zone", "zone list", zoneList)
	err = oceanbaseOperationManager.AlterPool(&model.PoolParam{
		PoolName: oceanbaseconst.SysTenantPool,
		ZoneList: zoneList,
	})
	if err != nil {
		return errors.Wrapf(err, "Failed to modify sys pool's zone list to  %v", zoneList)
	}
	// add locality one by one
	sysTenant, err := oceanbaseOperationManager.GetTenantByName(oceanbaseconst.SysTenant)
	if err != nil {
		return errors.Wrap(err, "Failed to get sys tenant info")
	}
	locality := sysTenant.Locality
	replicas := obutil.ConvertFromLocalityStr(locality)
	for _, zone := range desiredZones {
		found := false
		for _, r := range replicas {
			if zone == r.Zone {
				found = true
				break
			}
		}
		if !found {
			replicas = append(replicas, model.Replica{
				Type: oceanbaseconst.FullType,
				Num:  1,
				Zone: zone,
			})
			locality = obutil.ConvertToLocalityStr(replicas)
			m.Logger.Info("Modify sys tenant's locality when add zone", "locality", locality)
			err = oceanbaseOperationManager.SetTenant(model.TenantSQLParam{
				TenantName: oceanbaseconst.SysTenant,
				Locality:   locality,
			})
			if err != nil {
				return errors.Wrapf(err, "Failed to set sys locality to %s", locality)
			}
			err = oceanbaseOperationManager.WaitTenantLocalityChangeFinished(oceanbaseconst.SysTenant, oceanbaseconst.LocalityChangeTimeoutSeconds)
			if err != nil {
				return errors.Wrapf(err, "Locality change to %s not finished after timeout", locality)
			}
		}
	}
	// delete locality one by one
	for _, r := range replicas {
		found := false
		for _, zone := range desiredZones {
			if zone == r.Zone {
				found = true
				break
			}
		}
		if !found {
			newReplicas := obutil.OmitZoneFromReplicas(replicas, r.Zone)
			locality = obutil.ConvertToLocalityStr(newReplicas)
			m.Logger.Info("Modify sys tenant's locality when delete zone", "locality", locality)
			err = oceanbaseOperationManager.SetTenant(model.TenantSQLParam{
				TenantName: oceanbaseconst.SysTenant,
				Locality:   locality,
			})
			if err != nil {
				return errors.Wrapf(err, "Failed to set sys locality to %s", locality)
			}
			err = oceanbaseOperationManager.WaitTenantLocalityChangeFinished(oceanbaseconst.SysTenant, oceanbaseconst.LocalityChangeTimeoutSeconds)
			if err != nil {
				return errors.Wrapf(err, "Locality change to %s not finished after timeout", locality)
			}
		}
	}
	// delete zone from pool zone list
	newZoneList := make([]string, 0)
	for _, zone := range zoneList {
		found := false
		for _, z := range desiredZones {
			if zone == z {
				found = true
				break
			}
		}
		if found {
			newZoneList = append(newZoneList, zone)
		}
	}
	m.Logger.Info("Modify sys pool's zone list when delete zone", "zone list", newZoneList)
	return oceanbaseOperationManager.AlterPool(&model.PoolParam{
		PoolName: oceanbaseconst.SysTenantPool,
		ZoneList: newZoneList,
	})
}

func (m *OBClusterManager) CreateServiceForMonitor() tasktypes.TaskError {
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBCluster.APIVersion,
		Kind:       m.OBCluster.Kind,
		Name:       m.OBCluster.Name,
		UID:        m.OBCluster.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	selector := make(map[string]string)
	selector[oceanbaseconst.LabelRefOBCluster] = m.OBCluster.Name
	parts := strings.Split(uuid.New().String(), "-")
	suffix := parts[len(parts)-1]
	monitorService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       m.OBCluster.Namespace,
			Name:            fmt.Sprintf("svc-monitor-%s-%s", m.OBCluster.Name, suffix),
			OwnerReferences: ownerReferenceList,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       obagentconst.HttpPortName,
					Protocol:   corev1.ProtocolTCP,
					Port:       obagentconst.HttpPort,
					TargetPort: intstr.FromInt(obagentconst.HttpPort),
				},
			},
			Selector: selector,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	err := m.Client.Create(m.Ctx, &monitorService)
	if err != nil {
		return errors.Wrap(err, "Create monitor service")
	}
	m.Recorder.Event(m.OBCluster, "MaintainedAfterBootstrap", "", "Create monitor service successfully")
	return nil
}

func (m *OBClusterManager) RestoreEssentialParameters() tasktypes.TaskError {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	essentialParameters := make([]model.Parameter, 0)

	contextObjectName := fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, oceanbaseconst.EssentialParametersKey)
	contextSecret := &corev1.Secret{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBCluster.Namespace,
		Name:      contextObjectName,
	}, contextSecret)
	if err != nil {
		m.Logger.Error(err, "Failed to get context secret")
		// parameter can be set manually, just return here and emit an event
		m.Recorder.Event(m.OBCluster, "Warning", "Restore essential parameters failed", err.Error())
		return nil
	}

	encodedParameters := string(contextSecret.Data[oceanbaseconst.EssentialParametersKey])
	m.Logger.Info("Get encoded parameters", "parameters", encodedParameters)
	err = json.Unmarshal([]byte(encodedParameters), &essentialParameters)
	if err != nil {
		return errors.New("Parse encoded parameters failed")
	}

	for _, parameter := range essentialParameters {
		err = oceanbaseOperationManager.SetParameter(parameter.Name, parameter.Value, &param.Scope{
			Name:  "server",
			Value: fmt.Sprintf("%s:%d", parameter.SvrIp, parameter.SvrPort),
		})
		if err != nil {
			return errors.Wrapf(err, "Failed to set parameter %s to %s:%d", parameter.Name, parameter.SvrIp, parameter.SvrPort)
		}
	}
	_ = m.Client.Delete(m.Ctx, contextSecret)
	m.Recorder.Event(m.OBCluster, "Upgrade", "", "Restore essential parameters successfully")
	return nil
}

func (m *OBClusterManager) CheckAndCreateUserSecrets() tasktypes.TaskError {
	secretList := []string{
		m.OBCluster.Spec.UserSecrets.Operator,
		m.OBCluster.Spec.UserSecrets.Monitor,
		m.OBCluster.Spec.UserSecrets.ProxyRO,
	}
	for _, secret := range secretList {
		fetchedSec := &corev1.Secret{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBCluster.Namespace,
			Name:      secret,
		}, fetchedSec)
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				err := m.Client.Create(m.Ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secret,
						Namespace: m.OBCluster.Namespace,
					},
					StringData: map[string]string{
						"password": rand.String(16),
					},
				})
				if err != nil {
					return errors.Wrap(err, "Create secret "+secret)
				}
			}
		}
	}
	return nil
}

func (m *OBClusterManager) CreateServices() tasktypes.TaskError {
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist && modeAnnoVal == oceanbaseconst.ModeStandalone {
		err := m.Client.Create(m.Ctx, &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.OBCluster.GetName() + "-standalone-svc",
				Namespace: m.OBCluster.GetNamespace(),
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: m.OBCluster.APIVersion,
					Kind:       m.OBCluster.Kind,
					Name:       m.OBCluster.GetName(),
					UID:        m.OBCluster.GetUID(),
				}},
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Name:       "sql",
					Protocol:   corev1.ProtocolTCP,
					Port:       2881,
					TargetPort: intstr.IntOrString{IntVal: 2881},
				}},
				Selector: map[string]string{
					oceanbaseconst.LabelRefOBCluster: m.OBCluster.GetName(),
				},
				Type: corev1.ServiceTypeNodePort,
			},
		})
		if err != nil {
			m.Recorder.Event(m.OBCluster, "Warning", "Create standalone service failed", err.Error())
			return errors.Wrap(err, "Create service")
		}
	}
	return nil
}

func (m *OBClusterManager) CheckImageReady() tasktypes.TaskError {
	jobName := "image-pull-ready-" + rand.String(8)
	var ttl int32 = 120
	var backoffLimit int32 = 32
	checkImagePullJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: m.OBCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       m.OBCluster.Kind,
				APIVersion: m.OBCluster.APIVersion,
				Name:       m.OBCluster.Name,
				UID:        m.OBCluster.UID,
			}},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "helper-check-image-pull-ready",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Image:           m.OBCluster.Spec.OBServerTemplate.Image,
						Command:         []string{"bash", "-c", "/home/admin/oceanbase/bin/oceanbase-helper help"},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoffLimit,
		},
	}
	err := m.Client.Create(m.Ctx, checkImagePullJob)
	if err != nil {
		return errors.Wrap(err, "Create check image pull job")
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Create check image pull job", "job", jobName)

	imagePullReady := false
	var checkImagePullReadyMaxTimes = 8000
	labelSelector := metav1.FormatLabelSelector(checkImagePullJob.Spec.Selector)
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		return errors.Wrap(err, "Parse label selector")
	}
outerLoop:
	for i := 0; i < checkImagePullReadyMaxTimes; i++ {
		podList := &corev1.PodList{}
		err = m.Client.List(m.Ctx, podList, &client.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return errors.Wrap(err, "List pods")
		}
		if len(podList.Items) == 0 {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("No pod found for check image pull job")
			time.Sleep(time.Second * oceanbaseconst.CheckJobInterval)
			continue
		}
		pod := podList.Items[0]
		switch pod.Status.Phase {
		case corev1.PodFailed:
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Check image pull job failed")
			return errors.New("Check image pull job failed")
		case corev1.PodSucceeded, corev1.PodRunning:
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Check image pull job finished")
			imagePullReady = true
			break outerLoop
		case corev1.PodPending, corev1.PodUnknown:
			// if every container has pulled its image, break outer loop
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					switch containerStatus.State.Waiting.Reason {
					case "ErrImagePull", "ImagePullBackOff":
						m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Wait to pull image", "reason", containerStatus.State.Waiting.Reason, "message", containerStatus.State.Waiting.Message)
					default:
						m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Container is waiting", "reason", containerStatus.State.Waiting.Reason, "message", containerStatus.State.Waiting.Message)
					}
					time.Sleep(time.Second * oceanbaseconst.CheckJobInterval)
					continue outerLoop
				} else if containerStatus.State.Running != nil || containerStatus.State.Terminated != nil {
					m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Container is running or terminated")
				}
			}
			imagePullReady = true
			break outerLoop
		}
	}
	if !imagePullReady {
		return errors.New("Image pull not ready")
	}
	return nil
}

func (m *OBClusterManager) CheckClusterMode() tasktypes.TaskError {
	var err error
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBCluster, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist && modeAnnoVal == oceanbaseconst.ModeStandalone {
		var backoffLimit int32
		var ttl int32 = 300
		jobName := "standalone-validate-" + rand.String(8)
		standaloneValidateJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: m.OBCluster.Namespace,
				OwnerReferences: []metav1.OwnerReference{{
					Kind:       m.OBCluster.Kind,
					APIVersion: m.OBCluster.APIVersion,
					Name:       m.OBCluster.Name,
					UID:        m.OBCluster.UID,
				}},
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:    "helper-validate-standalone",
							Image:   m.OBCluster.Spec.OBServerTemplate.Image,
							Command: []string{"bash", "-c", "/home/admin/oceanbase/bin/oceanbase-helper standalone validate"},
						}},
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
				BackoffLimit:            &backoffLimit,
				TTLSecondsAfterFinished: &ttl,
			},
		}
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Create check version job", "job", jobName)

		err = m.Client.Create(m.Ctx, standaloneValidateJob)
		if err != nil {
			return errors.Wrap(err, "Create check version job")
		}

		var jobObject *batchv1.Job
		var maxCheckTimes = 600
		for i := 0; i < maxCheckTimes; i++ {
			time.Sleep(time.Second * oceanbaseconst.CheckJobInterval)
			jobObject, err = resourceutils.GetJob(m.Client, m.OBCluster.Namespace, jobName)
			if err != nil {
				m.Logger.Error(err, "Failed to get job")
				return err
			}
			if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("OBServer version check job is still running")
			} else {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("OBServer version check job finished")
				break
			}
		}
		if jobObject.Status.Failed > 0 {
			m.Logger.Info("Current image does not support standalone mode")
			err := errors.New("Current image does not support standalone mode")
			m.PrintErrEvent(err)
			return err
		}
	}
	return nil
}

// check obcluster name and id
// check obzone exists in topology
func (m *OBClusterManager) CheckMigration() tasktypes.TaskError {
	m.Logger.Info("Check before migration")
	manager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "get operation manager")
	}
	obzoneList, err := manager.ListZones()
	if err != nil {
		return errors.Wrap(err, "list obzones")
	}
	for _, obzone := range obzoneList {
		found := false
		for _, zone := range m.OBCluster.Spec.Topology {
			if obzone.Name == zone.Zone {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("obzone %s not in obcluster's topology", obzone.Name)
		}
	}
	obclusterNameParamList, err := manager.GetParameter(oceanbaseconst.ClusterNameParam, nil)
	if err != nil {
		return errors.Wrap(err, "get obcluster name failed")
	}
	obclusterName := obclusterNameParamList[0].Value
	obclusterIdParamList, err := manager.GetParameter(oceanbaseconst.ClusterIdParam, nil)
	if err != nil {
		return errors.Wrap(err, "get obcluster id failed")
	}
	obclusterId := obclusterIdParamList[0].Value
	if obclusterName != m.OBCluster.Spec.ClusterName {
		return errors.Errorf("Cluster name mismatch, source cluster: %s, current: %s", obclusterName, m.OBCluster.Spec.ClusterName)
	}
	if obclusterId != fmt.Sprintf("%d", m.OBCluster.Spec.ClusterId) {
		return errors.Errorf("Cluster id mismatch, source cluster: %s, current: %d", obclusterId, m.OBCluster.Spec.ClusterId)
	}
	return nil
}
