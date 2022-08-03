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

package converter

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	statefulappCore "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
	"github.com/pkg/errors"
)

func IsAllOBServerActive(obServerList []model.AllServer, obClusters []cloudv1.Cluster) bool {
	obServerCurrentReplicas := make(map[string]bool)
	for _, obServer := range obServerList {
		if obServer.Status == observerconst.OBServerActive && obServer.StartServiceTime > 0 {
			obServerCurrentReplicas[obServer.Zone] = true
		}
	}
	for _, obCluster := range obClusters {
		if obCluster.Cluster == myconfig.ClusterName {
			for _, zone := range obCluster.Zone {
				tmp := obServerCurrentReplicas[zone.Name]
				if !tmp {
					return false
				}
			}
			return true
		}
	}
	return false
}

func IsOBServerDeleted(clusterIP, podIP string) bool {
	obServerList := sql.GetOBServer(clusterIP)
	for _, obServer := range obServerList {
		if obServer.SvrIP == podIP {
			return false
		}
	}
	return true
}

func IsPodNotInOBServerList(zoneName, ip string, nodeMap map[string][]cloudv1.OBNode) bool {
	zoneIPList := nodeMap[zoneName]

	if len(zoneIPList) > 0 {
		for _, tmpIP := range zoneIPList {
			if tmpIP.ServerIP == ip {
				return false
			}
		}
		return true
	}
	return false
}

func IsPodInOBZoneListNotInOBServerList(zoneName string, nodeMap map[string][]cloudv1.OBNode, zoneNodeMap map[string][]cloudv1.OBZoneNode) bool {
	if len(nodeMap) > 0 {
		if nodeMap[zoneName] == nil && zoneNodeMap[zoneName] != nil {
			return true
		}
	}
	return false
}

func IsOBServerInactiveOrDeletingAndNotInPodList(server cloudv1.OBNode, podRunningList []string) bool {
	if server.Status == observerconst.OBServerInactive || server.Status == observerconst.OBServerDeleting {
		for _, podIP := range podRunningList {
			if podIP == server.ServerIP {
				return false
			}
		}
		return true
	}
	return false
}

func GetInfoForAddServerByZone(clusterIP string, statefulApp cloudv1.StatefulApp) (error, string, string) {
	obServerList := sql.GetOBServer(clusterIP)
	obZoneList := sql.GetOBZone(clusterIP)
	if len(obServerList) == 0 {
		return errors.New(observerconst.DataBaseError), "", ""
	}
	if len(obZoneList) == 0 {
		return errors.New(observerconst.DataBaseError), "", ""
	}
	nodeMap := GenerateNodeMapByOBServerList(obServerList)
	zoneNodeMap := GenerateZoneNodeMapByOBZoneList(obZoneList)

	// judge which ip need add
	for _, subset := range statefulApp.Status.Subsets {
		for _, pod := range subset.Pods {
			if pod.PodPhase == statefulappCore.PodStatusRunning && pod.Index < subset.ExpectedReplicas {
				status1 := IsPodNotInOBServerList(subset.Name, pod.PodIP, nodeMap)
				status2 := IsPodInOBZoneListNotInOBServerList(subset.Name, nodeMap, zoneNodeMap)
				// Pod IP not in OBServerList, need to add server
				// do one thing at a time
				if status1 || status2 {
					return nil, subset.Name, pod.PodIP
				}
			}
		}
	}

	return errors.New("none ip need add"), "", ""
}

func GetInfoForDelZone(clusterIP string, clusterSpec cloudv1.Cluster, statefulApp cloudv1.StatefulApp) (error, string) {
	obZoneList := sql.GetOBZone(clusterIP)
	if len(obZoneList) == 0 {
		return errors.New(observerconst.DataBaseError), ""
	}
	zoneNodeMap := GenerateZoneNodeMapByOBZoneList(obZoneList)

	for _, obZone := range obZoneList {
		zoneSpec := GetZoneSpecFromClusterSpec(obZone.Zone, clusterSpec)
		if zoneNodeMap[obZone.Zone] != nil && zoneSpec.Name == "" {
			return nil, obZone.Zone
		}
	}

	return errors.New("none zone need del"), ""
}

func GetInfoForDelServerByZone(clusterIP string, clusterSpec cloudv1.Cluster, statefulApp cloudv1.StatefulApp) (error, string, string) {
	obServerList := sql.GetOBServer(clusterIP)
	if len(obServerList) == 0 {
		return errors.New(observerconst.DataBaseError), "", ""
	}

	nodeMap := GenerateNodeMapByOBServerList(obServerList)

	// judge witch ip need del
	for _, subset := range statefulApp.Status.Subsets {
		podListToDelete := getPodListToDeleteFromSubsetStatus(subset)
		zoneSpec := GetZoneSpecFromClusterSpec(subset.Name, clusterSpec)
		// number of observer in db > replica
		if len(nodeMap[subset.Name]) > zoneSpec.Replicas {
			for _, pod := range nodeMap[subset.Name] {
				for _, podToDelete := range podListToDelete {
					if pod.ServerIP == podToDelete {
						return nil, subset.Name, pod.ServerIP
					}
				}
			}
		}
	}

	return errors.New("none ip need del"), "", ""
}

func getPodListToDeleteFromSubsetStatus(subset cloudv1.SubsetStatus) []string {
	podList := make([]string, 0)
	for _, pod := range subset.Pods {
		if pod.Index >= subset.ExpectedReplicas {
			podList = append(podList, pod.PodIP)
		}
	}
	return podList
}

func getRunningPodListFromSubsetStatus(subset cloudv1.SubsetStatus) []string {
	runningPodList := make([]string, 0)
	for _, pod := range subset.Pods {
		if pod.PodPhase == statefulappCore.PodStatusRunning {
			runningPodList = append(runningPodList, pod.PodIP)
		}
	}
	return runningPodList
}
