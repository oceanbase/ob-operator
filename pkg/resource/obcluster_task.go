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
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv2alpha1 "github.com/oceanbase/ob-operator/api/v2alpha1"
	zonestatus "github.com/oceanbase/ob-operator/pkg/const/status/obzone"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
)

func (m *OBClusterManager) getOBCluster() (*cloudv2alpha1.OBCluster, error) {
	obcluster := &cloudv2alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBCluster.Name), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBClusterManager) generateZoneName(zone string) string {
	return fmt.Sprintf("%s-%d-%s", m.OBCluster.Spec.ClusterName, m.OBCluster.Spec.ClusterId, zone)
}

func (m *OBClusterManager) FindTask(name string) (func() error, error) {
	switch name {
	case taskname.CreateOBZone:
		return m.CreateOBZone, nil
	case taskname.WaitOBZoneBootstrapReady:
		return m.WaitOBZoneBootstrapReady, nil
	case taskname.Bootstrap:
		return m.Bootstrap, nil
	case taskname.CreateUsers:
		return m.CreateUsers, nil
	case taskname.CreateOBParameter:
		return m.CreateOBParameter, nil
	default:
		return nil, errors.New(fmt.Sprintf("Can not find an function for %s", name))
	}
}

func (m *OBClusterManager) WaitOBZoneBootstrapReady() error {
	for i := 1; i < 300; i++ {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return errors.Wrap(err, "get obcluster failed")
		}
		allready := true
		for _, obzoneStatus := range obcluster.Status.OBZoneStatus {
			if obzoneStatus.Status != zonestatus.BootstrapReady {
				m.Logger.Info("zone still not ready for bootstrap", "zone", obzoneStatus.Zone)
				allready = false
				break
			}
		}
		if allready {
			return nil
		} else {
			time.Sleep(time.Second)
		}
	}
	return errors.New("all server still not bootstrap ready when timeout")
}

func (m *OBClusterManager) CreateOBZone() error {
	m.Logger.Info("create obzones")
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBCluster.APIVersion,
		Kind:       m.OBCluster.Kind,
		Name:       m.OBCluster.Name,
		UID:        m.OBCluster.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	for _, zone := range m.OBCluster.Spec.Topology {
		zoneName := m.generateZoneName(zone.Zone)
		labels := make(map[string]string)
		labels["reference-uid"] = string(m.OBCluster.GetUID())
		labels["reference-cluster"] = m.OBCluster.Name
		obzone := &cloudv2alpha1.OBZone{
			ObjectMeta: metav1.ObjectMeta{
				Name:            zoneName,
				Namespace:       m.OBCluster.Namespace,
				OwnerReferences: ownerReferenceList,
				Labels:          labels,
			},
			Spec: cloudv2alpha1.OBZoneSpec{
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

func (m *OBClusterManager) BootstrapEmpty() error {
	return nil
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
	address := obzoneList.Items[0].Status.OBServerStatus[0].Server
	p := connector.NewOceanbaseConnectProperties(address, 2881, "root", "sys", "root", "")
	manager, err := operation.GetOceanbaseOperationManager(p)
	if err != nil {
		m.Logger.Error(err, "get oceanbase sql operator failed")
		return errors.Wrap(err, "get oceanbase sql operator")
	}
	m.Logger.Info("successfully get oceanbase sql operator")

	bootstrapServers := make([]model.BootstrapServerInfo, 0, len(m.OBCluster.Spec.Topology))
	for _, zone := range obzoneList.Items {
		serverInfo := &model.ServerInfo{
			Ip:   zone.Status.OBServerStatus[0].Server,
			Port: 2882,
		}
		bootstrapServers = append(bootstrapServers, model.BootstrapServerInfo{
			Region: "default",
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
	return nil
}

func (m *OBClusterManager) CreateOBParameter() error {
	return nil
}
