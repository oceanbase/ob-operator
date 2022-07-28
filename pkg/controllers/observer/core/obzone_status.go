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
	"k8s.io/klog/v2"
	"reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
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
	klog.Infoln("UpdateOBZoneStatus: obZoneCurrent.Status ", obZoneCurrent.Status)
	klog.Infoln("UpdateOBZoneStatus: obZoneStatus.Status ", obZoneStatus.Status)
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
