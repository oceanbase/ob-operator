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

	cloudv2alpha1 "github.com/oceanbase/ob-operator/api/v2alpha1"
	serverstatus "github.com/oceanbase/ob-operator/pkg/const/status/observer"
)

func (m *OBZoneManager) generateServerName() string {
	parts := strings.Split(uuid.New().String(), "-")
	suffix := parts[len(parts)-1]
	return fmt.Sprintf("%s-%d-%s-%s", m.OBZone.Spec.ClusterName, m.OBZone.Spec.ClusterId, m.OBZone.Spec.Topology.Zone, suffix)
}

func (m *OBZoneManager) AddZone() error {
	return nil
}

func (m *OBZoneManager) StartZone() error {
	return nil
}

func (m *OBZoneManager) WaitOBServerBootstrapReady() error {
	for i := 1; i < 300; i++ {
		obzone, err := m.getOBZone()
		if err != nil {
			return errors.Wrap(err, "get obzoen failed")
		}
		allready := true
		for _, observerStatus := range obzone.Status.OBServerStatus {
			if observerStatus.Status != serverstatus.BootstrapReady {
				m.Logger.Info("server still not ready for bootstrap", "server ip", observerStatus.Server)
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

func (m *OBZoneManager) CreateOBServer() error {
	m.Logger.Info("create observers")
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBZone.APIVersion,
		Kind:       m.OBZone.Kind,
		Name:       m.OBZone.Name,
		UID:        m.OBZone.GetUID(),
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	for i := 0; i < m.OBZone.Spec.Topology.Replica; i++ {
		serverName := m.generateServerName()
		labels := make(map[string]string)
		cluster, _ := m.OBZone.Labels["reference-cluster"]
		labels["reference-uid"] = string(m.OBZone.GetUID())
		labels["reference-zone"] = m.OBZone.Name
		labels["reference-cluster"] = cluster
		observer := &cloudv2alpha1.OBServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:            serverName,
				Namespace:       m.OBZone.Namespace,
				OwnerReferences: ownerReferenceList,
				Labels:          labels,
			},
			Spec: cloudv2alpha1.OBServerSpec{
				ClusterName:      m.OBZone.Spec.ClusterName,
				ClusterId:        m.OBZone.Spec.ClusterId,
				Zone:             m.OBZone.Spec.Topology.Zone,
				NodeSelector:     m.OBZone.Spec.Topology.NodeSelector,
				OBServerTemplate: m.OBZone.Spec.OBServerTemplate,
				MonitorTemplate:  m.OBZone.Spec.MonitorTemplate,
				BackupVolume:     m.OBZone.Spec.BackupVolume,
			},
		}
		m.Logger.Info("create observer", "server", serverName)
		err := m.Client.Create(m.Ctx, observer)
		if err != nil {
			m.Logger.Error(err, "create observer failed", "server", serverName)
			return errors.Wrap(err, "create observer")
		}
	}
	return nil
}
