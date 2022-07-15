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
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/judge"
	"k8s.io/klog/v2"
)

func (ctrl *OBClusterCtrl) OBServerCoordinator(statefulApp cloudv1.StatefulApp) error {
	var err error
	scaleState, subset := judge.OBServerScale(ctrl.OBCluster.Spec.Topology, statefulApp)
	switch scaleState {
	case observerconst.ScaleUP:
		err = ctrl.UpdateOBServerReplica(subset, statefulApp, observerconst.ScaleUP)
	case observerconst.ScaleDown:
		err = ctrl.UpdateOBServerReplica(subset, statefulApp, observerconst.ScaleDown)
	case observerconst.Maintain:
		err = ctrl.OBServerMaintain(statefulApp)
	}
	return err
}

func (ctrl *OBClusterCtrl) UpdateOBServerReplica(subset cloudv1.Subset, statefulApp cloudv1.StatefulApp, status string) error {
	// generate new StatefulApp for new replica
	newStatefulApp := converter.UpdateSubsetReplicaForStatefulApp(subset, statefulApp)
	// update StatefulApp replica
	statefulAppCtrl := NewStatefulAppCtrl(ctrl, newStatefulApp)
	err := statefulAppCtrl.UpdateStatefulApp()
	if err != nil {
		return err
	}

	// generate new OBZone for new replica
	obZoneName := converter.GenerateOBZoneName(ctrl.OBCluster.Name)
	obZoneCtrl := NewOBZoneCtrl(ctrl)
	obZoneCurrent, err := obZoneCtrl.GetOBZoneByName(ctrl.OBCluster.Namespace, obZoneName)
	if err != nil {
		return err
	}
	newOBZone := converter.UpdateOBZoneSpec(obZoneCurrent, ctrl.OBCluster.Spec.Topology)
	// update OBZone replica
	err = obZoneCtrl.UpdateOBZone(newOBZone)
	if err != nil {
		return err
	}

	// update status
	return ctrl.UpdateOBClusterAndZoneStatus(status, "", "")
}

func (ctrl *OBClusterCtrl) OBServerScaleUPByZone(statefulApp cloudv1.StatefulApp) error {
	var clusterStatus string
	var zoneStatus string

	// get ClusterIP
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err != nil {
		return err
	}

	// get info for add server
	err, zoneName, podIP := converter.GetInfoForAddServerByZone(clusterIP, statefulApp)
	// nil need to add server
	if err == nil {
		clusterStatus = observerconst.ScaleUP
		// add server
		klog.Infoln("-----------------------OBServerScaleUPByZone-----------------------")
		err = ctrl.AddOBServer(clusterIP, zoneName, podIP, statefulApp)
		if err != nil {
			return err
		}
		zoneStatus = observerconst.OBServerAdd
		// update status
		return ctrl.UpdateOBClusterAndZoneStatus(clusterStatus, zoneName, zoneStatus)
	}

	// need fix status
	return ctrl.FixStatus()
}

func (ctrl *OBClusterCtrl) OBServerScaleDownByZone(statefulApp cloudv1.StatefulApp) error {
	var clusterStatus string
	var zoneStatus string

	// get ClusterIP
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err != nil {
		return err
	}

	// get info for del server
	clusterSpec := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	err, zoneName, podIP := converter.GetInfoForDelServerByZone(clusterIP, clusterSpec, statefulApp)
	// nil need to del server
	if err == nil {
		clusterStatus = observerconst.ScaleDown
		// del server
		err = ctrl.DelOBServer(clusterIP, zoneName, podIP)
		if err != nil {
			return err
		}
		zoneStatus = observerconst.OBServerDel
		// update status
		return ctrl.UpdateOBClusterAndZoneStatus(clusterStatus, zoneName, zoneStatus)
	}

	// need fix status
	return ctrl.FixStatus()
}

func (ctrl *OBClusterCtrl) OBServerMaintain(statefulApp cloudv1.StatefulApp) error {
	// get ClusterIP
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err != nil {
		return err
	}

	// get info for recover server
	err, zoneName, podIP := converter.GetInfoForRecoverServerByZone(clusterIP, statefulApp)
	// nil is need to recover server
	if err == nil {
		// add server
		klog.Info("need to recover server")
		err := ctrl.AsyncStartOBServer(clusterIP, zoneName, podIP, statefulApp)
		return err
	}

	// get info for add server
	err, zoneName, podIP = converter.GetInfoForAddServerByZone(clusterIP, statefulApp)
	// nil is need to add server
	if err == nil {
		// add server
		klog.Info("need to add server")
		return ctrl.AddOBServer(clusterIP, zoneName, podIP, statefulApp)
	}

	// get info for del server
	clusterSpec := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	err, zoneName, podIP = converter.GetInfoForDelServerByZone(clusterIP, clusterSpec, statefulApp)
	// nil need to del server
	if err == nil {
		// del server
		klog.Info("need to delete server")
		return ctrl.DelOBServer(clusterIP, zoneName, podIP)
	}

	return ctrl.OBClusterReadyForStep(observerconst.StepMaintain, statefulApp)
}

func (ctrl *OBClusterCtrl) FixStatus() error {
	var zoneName string
	var zoneStatus string
	var clusterStatus string
	oldClusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	for _, oldZoneStatus := range oldClusterStatus.Zone {
		if oldZoneStatus.ZoneStatus != observerconst.OBZoneReady {
			zoneName = oldZoneStatus.Name
			zoneStatus = observerconst.OBZoneReady
			clusterStatus = observerconst.ClusterReady
			break
		}
	}
	return ctrl.UpdateOBClusterAndZoneStatus(clusterStatus, zoneName, zoneStatus)
}
