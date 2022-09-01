/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package core

import (
	"context"
	"reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *OBClusterCtrl) UpdateOBZoneStatus(statefulApp cloudv1.StatefulApp) error {
	subsets := statefulApp.Status.Subsets
	// TODO: check owner
	obZoneName := converter.GenerateOBZoneName(ctrl.OBCluster.Name)
	obZoneCtrl := NewOBZoneCtrl(ctrl)
	obZoneCurrent, err := obZoneCtrl.GetOBZoneByName(ctrl.OBCluster.Namespace, obZoneName)
	if err != nil {
		return err
	}
	obServerList := sql.GetOBServer(subsets[0].Pods[0].PodIP)
	cluster := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	obZoneStatus := converter.OBServerListToOBZoneStatus(cluster, obZoneCurrent, obServerList)
	status := reflect.DeepEqual(obZoneCurrent.Status, obZoneStatus.Status)
	if !status {
		err = obZoneCtrl.UpdateOBZoneStatus(obZoneStatus)
		if err != nil {
			return err
		}
		kube.LogForAppActionStatus(obZoneStatus.Kind, obZoneName, "update status", obZoneStatus)
	}
	return nil
}

func (ctrl *OBClusterCtrl) buildOBZoneStatusFromDB(obCluster cloudv1.OBCluster, clusterIP string) (cloudv1.OBCluster, error) {
	expectedOBZoneList := ctrl.OBCluster.Spec.Topology[0].Zone
	obZoneListFromDB := sql.GetOBZone(clusterIP)

	// 期望的 zone 比实际的 少
	if len(expectedOBZoneList) < len(obZoneListFromDB) {
		obCluster.Status.Status = observerconst.TopologyNotReady
		return obCluster, nil
	}
	isOK := true
	for _, zone := range expectedOBZoneList {
		for _, obZoneByDB := range obZoneListFromDB {
			// 查询ob 看该zone 是否 active
			if zone.Name == obZoneByDB.Zone && obZoneByDB.Info != observerconst.OBZoneActive {
				isOK = false
			}
		}
	}
	if isOK {
		obCluster.Status.Status = observerconst.TopologyReady
	}
	return obCluster, nil

}

func (ctrl *OBClusterCtrl) UpdateOBZoneStatusFromDB(clusterIP string) error {
	obClusterNew, err := ctrl.buildOBZoneStatusFromDB(ctrl.OBCluster, clusterIP)
	obClusterExecuter := resource.NewOBClusterResource(ctrl.Resource)
	if err != nil {
		return err
	}

	err = obClusterExecuter.UpdateStatus(context.TODO(), obClusterNew)
	if err != nil {
		return err
	}

	return nil
}
