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
	"fmt"
	"sort"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	observerutil "github.com/oceanbase/ob-operator/pkg/controllers/observer/core/util"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
)

func GenerateOBZoneName(name string) string {
	statefulAppName := fmt.Sprintf("obzone-%s", name)
	return statefulAppName
}

func GenerateOBZoneSpec(obCluster cloudv1.OBCluster) cloudv1.OBZoneSpec {
	spec := cloudv1.OBZoneSpec{
		Topology: obCluster.Spec.Topology,
	}
	return spec
}

func GenerateOBZoneObject(obCluster cloudv1.OBCluster) cloudv1.OBZone {
	name := GenerateOBZoneName(obCluster.Name)
	objectMeta := observerutil.GenerateObjectMeta(obCluster, name)
	spec := GenerateOBZoneSpec(obCluster)
	obZone := cloudv1.OBZone{
		ObjectMeta: objectMeta,
		Spec:       spec,
	}
	return obZone
}

func GenerateOBZoneInfoListByCluster(cluster cloudv1.Cluster, nodeMap map[string][]cloudv1.OBNode) []cloudv1.OBZoneInfo {
	zoneList := make([]cloudv1.OBZoneInfo, 0)
	for _, zone := range cluster.Zone {
		v := make([]cloudv1.OBNode, 0, 0)
		vFromNodeMap := nodeMap[zone.Name]
		if nil != vFromNodeMap {
			v = vFromNodeMap
		}
		var zoneTemp cloudv1.OBZoneInfo
		zoneTemp.Name = zone.Name
		zoneTemp.Nodes = v
		zoneList = append(zoneList, zoneTemp)
	}
	return zoneList
}

func GenerateMultiClusterOBZoneStatus(obZoneCurrent cloudv1.OBZone, clusterOBZoneStatus cloudv1.ClusterOBZoneStatus) []cloudv1.ClusterOBZoneStatus {
	obZoneTopologyStatus := make([]cloudv1.ClusterOBZoneStatus, 0)
	if len(obZoneCurrent.Status.Topology) > 0 {
		for _, otherClusterStatus := range obZoneCurrent.Status.Topology {
			if otherClusterStatus.Cluster != myconfig.ClusterName {
				obZoneTopologyStatus = append(obZoneTopologyStatus, otherClusterStatus)
			}
		}
	}
	obZoneTopologyStatus = append(obZoneTopologyStatus, clusterOBZoneStatus)
	return obZoneTopologyStatus
}

func GenerateNodeMapByOBServerList(obServerList []model.AllServer) map[string][]cloudv1.OBNode {
	nodeMap := make(map[string][]cloudv1.OBNode, 0)
	for _, server := range obServerList {
		nodes := nodeMap[server.Zone]
		var node cloudv1.OBNode
		node.ServerIP = server.SvrIP
		node.Status = server.Status
		nodes = append(nodes, node)
		tmp := SortNodesStatus(nodes)
		nodeMap[server.Zone] = tmp
	}
	return nodeMap
}

func GenerateZoneNodeMapByOBZoneList(obZoneList []model.AllZone) map[string][]cloudv1.OBZoneNode {
	nodeMap := make(map[string][]cloudv1.OBZoneNode, 0)
	for _, zone := range obZoneList {
		nodes := nodeMap[zone.Zone]
		var node cloudv1.OBZoneNode
		node.Name = zone.Name
		node.Status = zone.Info
		nodes = append(nodes, node)
		nodeMap[zone.Zone] = nodes
	}
	return nodeMap
}

func UpdateOBZoneSpec(obzone cloudv1.OBZone, topology []cloudv1.Cluster) cloudv1.OBZone {
	obzone.Spec.Topology = topology
	return obzone
}

func OBServerListToOBZoneStatus(cluster cloudv1.Cluster, obZoneCurrent cloudv1.OBZone, obServerList []model.AllServer) cloudv1.OBZone {
	nodeMap := GenerateNodeMapByOBServerList(obServerList)

	zoneList := GenerateOBZoneInfoListByCluster(cluster, nodeMap)

	var clusterOBZoneStatus cloudv1.ClusterOBZoneStatus
	clusterOBZoneStatus.Cluster = myconfig.ClusterName
	clusterOBZoneStatus.Zone = zoneList

	// multi cluster
	obZoneTopologyStatus := GenerateMultiClusterOBZoneStatus(obZoneCurrent, clusterOBZoneStatus)

	obZoneCurrent.Status.Topology = obZoneTopologyStatus
	return obZoneCurrent
}

func SortNodesStatus(nodesStatus []cloudv1.OBNode) []cloudv1.OBNode {
	sort.Slice(nodesStatus, func(i, j int) bool {
		if nodesStatus[i].ServerIP < nodesStatus[j].ServerIP {
			return true
		}
		return false
	})
	return nodesStatus
}

func GetZoneSpecFromClusterSpec(zoneName string, clusterSpec cloudv1.Cluster) cloudv1.Subset {
	var res cloudv1.Subset
	for _, zone := range clusterSpec.Zone {
		if zone.Name == zoneName {
			res = zone
			break
		}
	}
	return res
}
