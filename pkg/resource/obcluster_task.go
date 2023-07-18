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

	obagentconst "github.com/oceanbase/ob-operator/pkg/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
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
		zoneDeleted := true
		for _, zoneStatus := range m.OBCluster.Status.OBZoneStatus {
			found := false
			for _, zone := range m.OBCluster.Spec.Topology {
				if zoneStatus.Zone == zone.Zone {
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
}

func (m *OBClusterManager) DeleteOBZone() error {
	obzoneList, err := m.listOBZones()
	if err != nil {
		m.Logger.Error(err, "List obzone failed")
		return errors.Wrapf(err, "List obzone of obcluster %s failed", m.OBCluster.Name)
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
			err = m.Client.Delete(m.Ctx, &obzone)
			if err != nil {
				return errors.Wrapf(err, "Delete obzone %s", obzone.Name)
			}
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
	return GetOceanbaseOperationManagerFromOBCluster(m.Client, m.OBCluster)
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
	manager, err := GetOceanbaseOperationManagerFromOBCluster(m.Client, m.OBCluster)
	if err != nil {
		m.Logger.Error(err, "get oceanbase operation manager failed")
		return errors.Wrap(err, "get oceanbase operation manager")
	}

	bootstrapServers := make([]model.BootstrapServerInfo, 0, len(m.OBCluster.Spec.Topology))
	for _, zone := range obzoneList.Items {
		serverInfo := &model.ServerInfo{
			Ip:   zone.Status.OBServerStatus[0].Server,
			Port: oceanbaseconst.RpcPort,
		}
		bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
			Region: oceanbaseconst.DefaultRegion,
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
