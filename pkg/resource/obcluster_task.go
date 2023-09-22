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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	obagentconst "github.com/oceanbase/ob-operator/pkg/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	zonestatus "github.com/oceanbase/ob-operator/pkg/const/status/obzone"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/param"
	obutil "github.com/oceanbase/ob-operator/pkg/oceanbase/util"
)

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
	return fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, strings.Replace(name, "_", "-", -1))
}

func (m *OBClusterManager) WaitOBZoneTopologyMatch() error {
	// TODO
	return nil
}

func (m *OBClusterManager) WaitOBZoneDeleted() error {
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
			m.Logger.Info("All zone deleted")
			waitSuccess = true
			break
		}
		time.Sleep(time.Second * 1)
	}
	if waitSuccess {
		return nil
	} else {
		return errors.Errorf("OBCluster %s zone still not deleted when timeout", m.OBCluster.Name)
	}
}

func (m *OBClusterManager) generateWaitOBZoneStatusFunc(status string, timeoutSeconds int) func() error {
	f := func() error {
		for i := 1; i < timeoutSeconds; i++ {
			obcluster, err := m.getOBCluster()
			if err != nil {
				return errors.Wrap(err, "get obcluster failed")
			}
			allMatched := true
			for _, obzoneStatus := range obcluster.Status.OBZoneStatus {
				if obzoneStatus.Status != status {
					m.Logger.Info("zone status still not matched", "zone", obzoneStatus.Zone, "status", status)
					allMatched = false
					break
				}
			}
			if allMatched {
				return nil
			}
			time.Sleep(time.Second)
		}
		return errors.New("zone status still not matched when timeout")
	}
	return f
}

func (m *OBClusterManager) ModifyOBZoneReplica() error {
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
			m.Logger.Info("Need to delete obzone", "obzone", obzone.Name)
			deletedZones = append(deletedZones, obzone)
		}
	}
	return deletedZones, nil
}

func (m *OBClusterManager) DeleteOBZone() error {
	zonesToDelete, err := m.getZonesToDelete()
	if err != nil {
		return errors.Wrap(err, "Failed to get obzones to delete")
	}
	for _, zone := range zonesToDelete {
		err = m.Client.Delete(m.Ctx, &zone)
		if err != nil {
			return errors.Wrapf(err, "Delete obzone %s", zone.Name)
		}
	}
	return nil
}

func (m *OBClusterManager) CreateOBZone() error {
	m.Logger.Info("create obzones")
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
			},
		}
		m.Logger.Info("create obzone", "zone", zoneName)
		err := m.Client.Create(m.Ctx, obzone)
		if err != nil {
			m.Logger.Error(err, "create obzone failed", "zone", zone.Zone)
			return errors.Wrap(err, "create obzone")
		}
	}
	return nil
}

func (m *OBClusterManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	return GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, m.OBCluster)
}

func (m *OBClusterManager) Bootstrap() error {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "list obzones failed")
		return errors.Wrap(err, "list obzones")
	}
	m.Logger.Info("successfully get obzone list", "obzone list", obzoneList)
	if len(obzoneList.Items) <= 0 {
		return errors.Wrap(err, "no obzone belongs to this cluster")
	}
	var manager *operation.OceanbaseOperationManager
	for i := 0; i < oceanbaseconst.GetConnectionMaxRetries; i++ {
		manager, err = m.getOceanbaseOperationManager()
		if err != nil || manager == nil {
			m.Logger.Info("Get oceanbase operation manager failed")
			time.Sleep(time.Second * oceanbaseconst.CheckConnectionInterval)
		} else {
			m.Logger.Info("Successfully got oceanbase operation manager")
			break
		}
	}
	if err != nil {
		m.Logger.Error(err, "get oceanbase operation manager failed")
		return errors.Wrap(err, "get oceanbase operation manager")
	}

	bootstrapServers := make([]model.BootstrapServerInfo, 0, len(m.OBCluster.Spec.Topology))
	connectAddress := manager.Connector.DataSource().GetAddress()
	for _, zone := range obzoneList.Items {
		serverIp := zone.Status.OBServerStatus[0].Server
		for _, serverInfo := range zone.Status.OBServerStatus {
			if serverInfo.Server == connectAddress {
				serverIp = connectAddress
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

	err = manager.Bootstrap(bootstrapServers)
	if err != nil {
		m.Logger.Error(err, "bootstrap failed")
	}
	return err
}

func (m *OBClusterManager) CreateService() error {
	return nil
}

func (m *OBClusterManager) CreateUsers() error {
	err := m.createUser(oceanbaseconst.OperatorUser, m.OBCluster.Spec.UserSecrets.Operator, oceanbaseconst.AllPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create operator user")
	}
	err = m.createUser(obagentconst.MonitorUser, m.OBCluster.Spec.UserSecrets.Monitor, oceanbaseconst.SelectPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	err = m.createUser(oceanbaseconst.ProxyUser, m.OBCluster.Spec.UserSecrets.ProxyRO, oceanbaseconst.SelectPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	err = m.createUser(oceanbaseconst.RootUser, m.OBCluster.Spec.UserSecrets.Root, oceanbaseconst.AllPrivilege)
	if err != nil {
		return errors.Wrap(err, "Create root user")
	}
	return nil
}

func (m *OBClusterManager) createUser(userName, secretName, privilege string) error {
	m.Logger.Info("begin create user", "username", userName)
	password, err := ReadPassword(m.Client, m.OBCluster.Namespace, secretName)
	if err != nil {
		return errors.Wrapf(err, "Get password from secret %s failed", secretName)
	}
	m.Logger.Info("finish get password", "username", userName, "password", password)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	m.Logger.Info("finish get operationmanager", "username", userName)
	err = oceanbaseOperationManager.CreateUser(userName)
	if err != nil {
		m.Logger.Error(err, "Create user")
		return errors.Wrapf(err, "Create user %s", userName)
	}
	m.Logger.Info("finish create user", "username", userName)
	err = oceanbaseOperationManager.SetUserPassword(userName, password)
	if err != nil {
		m.Logger.Error(err, "Set user password")
		return errors.Wrapf(err, "Set password for user %s", userName)
	}
	m.Logger.Info("finish set user password", "username", userName)
	object := "*.*"
	err = oceanbaseOperationManager.GrantPrivilege(privilege, object, userName)
	if err != nil {
		m.Logger.Error(err, "Grant privilege")
		return errors.Wrapf(err, "Grant privilege for user %s", userName)
	}
	m.Logger.Info("finish grant user privilege", "username", userName)
	return nil
}

func (m *OBClusterManager) MaintainOBParameter() error {
	parameterMap := make(map[string]v1alpha1.Parameter)
	for _, parameter := range m.OBCluster.Status.Parameters {
		m.Logger.Info("Build parameter map", "parameter", parameter.Name)
		parameterMap[parameter.Name] = parameter
	}
	for _, parameter := range m.OBCluster.Spec.Parameters {
		parameterStatus, parameterExists := parameterMap[parameter.Name]
		if !parameterExists {
			m.Logger.Info("Parameter not exists, need create", "param", parameter.Name)
			err := m.CreateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Crate obparameter failed", "param", parameter.Name)
			}
		} else if parameterStatus.Value != parameter.Value {
			m.Logger.Info("Parameter value not matched, need update", "param", parameter.Name)
			err := m.UpdateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Update obparameter failed", "param", parameter.Name)
			}
		}
		m.Logger.Info("Remove parameter from map", "parameter", parameter.Name)
		delete(parameterMap, parameter.Name)
	}

	// delete parameters that not in spec definition
	for _, parameter := range parameterMap {
		m.Logger.Info("Delete parameter", "parameter", parameter.Name)
		err := m.DeleteOBParameter(&parameter)
		if err != nil {
			m.Logger.Error(err, "Failed to delete parameter")
		}
	}
	return nil
}

func (m *OBClusterManager) CreateOBParameter(parameter *v1alpha1.Parameter) error {
	m.Logger.Info("create ob parameters")
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
	m.Logger.Info("create obparameter", "parameter", parameterName)
	err := m.Client.Create(m.Ctx, obparameter)
	if err != nil {
		m.Logger.Error(err, "create obparameter failed")
		return errors.Wrap(err, "create obparameter")
	}
	return nil
}

func (m *OBClusterManager) UpdateOBParameter(parameter *v1alpha1.Parameter) error {
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

func (m *OBClusterManager) DeleteOBParameter(parameter *v1alpha1.Parameter) error {
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

func (m *OBClusterManager) ValidateUpgradeInfo() error {
	// Get current obcluster version
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Failed to get operation manager of obcluster %s", m.OBCluster.Name)
	}
	//version, err := oceanbaseOperationManager.GetVersion()
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

	m.Logger.Info("Create validate upgrade job", "job", jobName)
	err = m.Client.Create(m.Ctx, &job)
	if err != nil {
		return errors.Wrapf(err, "Failed to create validate job for obcluster %s", m.OBCluster.Name)
	}

	var jobObject *batchv1.Job
	for {
		jobObject, err = GetJob(m.Client, m.OBCluster.Namespace, jobName)
		if err != nil {
			m.Logger.Error(err, "Failed to get job")
		}
		if jobObject.Status.Succeeded == 0 && jobObject.Status.Failed == 0 {
			m.Logger.Info("job is still running")
		} else {
			m.Logger.Info("job finished")
			break
		}
	}
	if jobObject.Status.Succeeded == 1 {
		m.Logger.Info("job succeeded")
	} else {
		m.Logger.Info("job is failed", "job", jobName)
		return errors.Wrap(err, "Failed to run validate job")
	}
	return nil
}

func (m *OBClusterManager) UpgradeCheck() error {
	return ExecuteUpgradeScript(m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradeCheckerScriptPath, "")
}

func (m *OBClusterManager) BackupEssentialParameters() error {
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

func (m *OBClusterManager) BeginUpgrade() error {
	return ExecuteUpgradeScript(m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradePreScriptPath, "")
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
			m.Logger.Info("Obzone upgrade finished", "obzone", zoneName)
			break
		}
		time.Sleep(time.Second * oceanbaseconst.CommonCheckInterval)
	}
	return nil
}

// TODO: add timeout
func (m *OBClusterManager) RollingUpgradeByZone() error {
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

func (m *OBClusterManager) FinishUpgrade() error {
	return ExecuteUpgradeScript(m.Client, m.Logger, m.OBCluster, oceanbaseconst.UpgradePostScriptPath, "")
}

func (m *OBClusterManager) ModifySysTenantReplica() error {
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
	m.Logger.Info("modify sys pool's zone list when add zone", "zone list", zoneList)
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
			m.Logger.Info("modify sys tenant's locality when add zone", "locality", locality)
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
			m.Logger.Info("modify sys tenant's locality when delete zone", "locality", locality)
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
	m.Logger.Info("modify sys pool's zone list when delete zone", "zone list", newZoneList)
	return oceanbaseOperationManager.AlterPool(&model.PoolParam{
		PoolName: oceanbaseconst.SysTenantPool,
		ZoneList: newZoneList,
	})
}

func (m *OBClusterManager) CreateServiceForMonitor() error {
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
	return m.Client.Create(m.Ctx, &monitorService)
}

func (m *OBClusterManager) RestoreEssentialParameters() error {
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
		return nil
		// parameter can be set manually, just return here and emit an event
		// TODO: emit an event
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
	m.Client.Delete(m.Ctx, contextSecret)
	return nil
}
