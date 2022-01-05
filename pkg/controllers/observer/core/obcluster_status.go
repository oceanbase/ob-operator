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
	"runtime"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	observerutil "github.com/oceanbase/ob-operator/pkg/controllers/observer/core/util"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *OBClusterCtrl) OBClusterReadyForStep(step string, statefulApp cloudv1.StatefulApp) error {
	// update RootService
	err := ctrl.UpdateRootServiceStatus(statefulApp)
	if err != nil {
		return err
	}

	// update OBZone
	err = ctrl.UpdateOBZoneStatus(statefulApp)
	if err != nil {
		return err
	}

	// create service
	switch step {
	case observerconst.StepBootstrap:
		err = ctrl.CreateService(statefulApp.Name)
		if err != nil {
			return err
		}
	case observerconst.StepMaintain:
		_, err = ctrl.GetServiceByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
		if err != nil {
			err = ctrl.CreateService(statefulApp.Name)
			if err != nil {
				return err
			}
		}
	}

	// update status
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
}

func (ctrl *OBClusterCtrl) UpdateOBClusterAndZoneStatus(clusterStatus, zoneName, zoneStatus string) error {
	var compareStatus bool
	obCluster := ctrl.OBCluster
	obClusterExecuter := resource.NewOBClusterResource(ctrl.Resource)
	// use retry to update
	retryErr := retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {
			// get current OBCluster every time
			obClusterTemp, err := obClusterExecuter.Get(context.TODO(), obCluster.Namespace, obCluster.Name)
			if err != nil {
				return err
			}
			// DeepCopy
			obClusterCurrent := obClusterTemp.(cloudv1.OBCluster)
			obClusterCurrentDeepCopy := obClusterCurrent.DeepCopy()
			// assign a value
			ctrl.OBCluster = *obClusterCurrentDeepCopy
			// build status
			obClusterNew, err := ctrl.buildOBClusterStatus(*obClusterCurrentDeepCopy, clusterStatus, zoneName, zoneStatus)
			if err != nil {
				return err
			}
			// compare status, if Equal don't need update
			compareStatus = reflect.DeepEqual(obClusterCurrent.Status, obClusterNew.Status)
			if !compareStatus {
				// update status
				err = obClusterExecuter.UpdateStatus(context.TODO(), obClusterNew)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
	if retryErr != nil {
		klog.Errorln(retryErr)
		return retryErr
	}
	// log
	if !compareStatus {
		p, _, _, _ := runtime.Caller(1)
		tmp := strings.Split(runtime.FuncForPC(p).Name(), "/")
		funcName := tmp[len(tmp)-1]
		observerutil.LogForOBClusterStatusConvert(funcName, ctrl.OBCluster.Name, clusterStatus, zoneName, zoneStatus)
	}
	return nil
}

func (ctrl *OBClusterCtrl) buildOBClusterStatus(obCluster cloudv1.OBCluster, clusterStatus, zoneName, zoneStatus string) (cloudv1.OBCluster, error) {
	statefulAppName := converter.GenerateStatefulAppName(obCluster.Name)
	statefulApp := &cloudv1.StatefulApp{}
	statefulAppCtrl := NewStatefulAppCtrl(ctrl, *statefulApp)
	// TODO: check owner
	statefulAppCurrent, err := statefulAppCtrl.GetStatefulAppByName(statefulAppName)
	if err != nil {
		return obCluster, err
	}

	cluster := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)

	nodeMap := make(map[string][]cloudv1.OBNode)
	// get ClusterIP
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err == nil {
		// get nodeMap from DB
		nodeMap = getNodeMapFromDB(clusterIP)
	}

	zoneList := buildZoneStatusList(cluster, statefulAppCurrent, nodeMap, zoneName, zoneStatus)

	// old cluster status
	var lastTransitionTime metav1.Time
	oldClusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	// old cluster status != now cluster status, need update lastTransitionTime & status
	if oldClusterStatus.ClusterStatus != clusterStatus {
		lastTransitionTime = metav1.Now()
	} else {
		lastTransitionTime = oldClusterStatus.LastTransitionTime
	}

	// new cluster status
	var clusterCurrentStatus cloudv1.ClusterStatus
	clusterCurrentStatus.Cluster = myconfig.ClusterName
	clusterCurrentStatus.ClusterStatus = clusterStatus
	clusterCurrentStatus.LastTransitionTime = lastTransitionTime
	clusterCurrentStatus.Zone = zoneList

	// topology status, multi cluster
	topologyStatus := buildMultiClusterStatus(obCluster, clusterCurrentStatus)

	if clusterStatus == observerconst.ClusterReady {
		obCluster.Status.Status = observerconst.TopologyReady
	} else if clusterStatus == observerconst.ScaleUP || clusterStatus == observerconst.ScaleDown {
		obCluster.Status.Status = observerconst.TopologyNotReady
	} else {
		obCluster.Status.Status = observerconst.TopologyPrepareing
	}
	obCluster.Status.Topology = topologyStatus
	return obCluster, nil
}

func buildZoneStatusList(cluster cloudv1.Cluster, statefulAppCurrent cloudv1.StatefulApp, nodeMap map[string][]cloudv1.OBNode, name, status string) []cloudv1.ZoneStatus {
	zoneList := make([]cloudv1.ZoneStatus, 0)
	for _, zone := range cluster.Zone {
		zoneStatus := buildZoneStatus(statefulAppCurrent, nodeMap, name, status, zone)
		zoneList = append(zoneList, zoneStatus)
	}
	return zoneList
}

func buildZoneStatus(statefulAppCurrent cloudv1.StatefulApp, nodeMap map[string][]cloudv1.OBNode, name, status string, zone cloudv1.Subset) cloudv1.ZoneStatus {
	subsetStatus := converter.GetSubsetStatusFromStatefulApp(zone.Name, statefulAppCurrent)
	var zoneStatus cloudv1.ZoneStatus
	zoneStatus.Name = zone.Name
	zoneStatus.Region = zone.Region
	zoneStatus.ExpectedReplicas = zone.Replicas
	// real AvailableReplicas from OB
	nodeList := nodeMap[zone.Name]
	zoneStatus.AvailableReplicas = int32(len(nodeList))
	// StatefulApp is not ready
	if subsetStatus.ExpectedReplicas != subsetStatus.AvailableReplicas {
		zoneStatus.ZoneStatus = observerconst.OBZonePrepareing
	} else {
		if zoneStatus.ExpectedReplicas > zoneStatus.AvailableReplicas {
			zoneStatus.ZoneStatus = observerconst.ScaleUP
		} else if zoneStatus.ExpectedReplicas < zoneStatus.AvailableReplicas {
			zoneStatus.ZoneStatus = observerconst.ScaleDown
		} else {
			zoneStatus.ZoneStatus = observerconst.OBZoneReady
		}
	}
	// use custom status
	if name == zone.Name && status != "" {
		zoneStatus.ZoneStatus = status
	}
	return zoneStatus
}

func buildMultiClusterStatus(obCluster cloudv1.OBCluster, clusterCurrentStatus cloudv1.ClusterStatus) []cloudv1.ClusterStatus {
	topologyStatus := make([]cloudv1.ClusterStatus, 0)
	if len(obCluster.Status.Topology) > 0 {
		for _, otherClusterStatus := range obCluster.Status.Topology {
			if otherClusterStatus.Cluster != myconfig.ClusterName {
				topologyStatus = append(topologyStatus, otherClusterStatus)
			}
		}
	}
	topologyStatus = append(topologyStatus, clusterCurrentStatus)
	return topologyStatus
}

func getNodeMapFromDB(clusterIP string) map[string][]cloudv1.OBNode {
	nodeMap := make(map[string][]cloudv1.OBNode)
	obServerList := sql.GetOBServer(clusterIP)
	if len(obServerList) > 0 {
		nodeMap = converter.GenerateNodeMapByOBServerList(obServerList)
	}
	return nodeMap
}
