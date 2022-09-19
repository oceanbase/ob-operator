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
	statefulappCore "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
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

// TODO refactor the following 3 function
func GetPodListToDeleteFromSubsetStatus(subset cloudv1.SubsetStatus) []string {
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
