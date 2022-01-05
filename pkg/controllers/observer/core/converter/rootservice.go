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

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	observerutil "github.com/oceanbase/ob-operator/pkg/controllers/observer/core/util"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
)

func GenerateRootServiceName(name string) string {
	statefulAppName := fmt.Sprintf("rs-%s", name)
	return statefulAppName
}

func GenerateRootServiceSpec(obCluster cloudv1.OBCluster) cloudv1.RootServiceSpec {
	spec := cloudv1.RootServiceSpec{
		Topology: obCluster.Spec.Topology,
	}
	return spec
}

func GenerateRootServiceObject(obCluster cloudv1.OBCluster) cloudv1.RootService {
	name := GenerateRootServiceName(obCluster.Name)
	objectMeta := observerutil.GenerateObjectMeta(obCluster, name)
	spec := GenerateRootServiceSpec(obCluster)
	rootService := cloudv1.RootService{
		ObjectMeta: objectMeta,
		Spec:       spec,
	}
	return rootService
}

func GenerateZoneRSList(cluster cloudv1.Cluster, rsList []model.AllVirtualCoreMeta, obServerList []model.AllServer) []cloudv1.ZoneRootServiceStatus {
	zoneRSList := make([]cloudv1.ZoneRootServiceStatus, 0)
	for _, zone := range cluster.Zone {
		zoneRS := GenerateZoneRootServiceStatusByRSList(zone.Name, rsList, obServerList)
		zoneRSList = append(zoneRSList, zoneRS)
	}
	return zoneRSList
}

func GenerateZoneRootServiceStatusByRSList(zoneName string, rsList []model.AllVirtualCoreMeta, obServerList []model.AllServer) cloudv1.ZoneRootServiceStatus {
	var zrs cloudv1.ZoneRootServiceStatus
	for _, rs := range rsList {
		if rs.Zone == zoneName {
			zrs.Name = rs.Zone
			zrs.ServerIP = rs.SvrIP
			zrs.Role = int(rs.Role)
			for _, server := range obServerList {
				if rs.SvrIP == server.SvrIP {
					zrs.Status = server.Status
					break
				}
			}
			break
		}
	}
	return zrs
}

func GenerateMultiClusterRootServiceStatus(rootService cloudv1.RootService, rsCurrentStatus cloudv1.ClusterRootServiceStatus) []cloudv1.ClusterRootServiceStatus {
	rsTopologyStatus := make([]cloudv1.ClusterRootServiceStatus, 0)
	if len(rootService.Status.Topology) > 0 {
		for _, otherClusterStatus := range rootService.Status.Topology {
			if otherClusterStatus.Cluster != myconfig.ClusterName {
				rsTopologyStatus = append(rsTopologyStatus, otherClusterStatus)
			}
		}
	}
	rsTopologyStatus = append(rsTopologyStatus, rsCurrentStatus)
	return rsTopologyStatus
}

func RSListToRSStatus(cluster cloudv1.Cluster, rootService cloudv1.RootService, rsList []model.AllVirtualCoreMeta, obServerList []model.AllServer) cloudv1.RootService {
	zrsList := GenerateZoneRSList(cluster, rsList, obServerList)

	var rsCurrentStatus cloudv1.ClusterRootServiceStatus
	rsCurrentStatus.Cluster = myconfig.ClusterName
	rsCurrentStatus.Zone = zrsList

	// multi cluster
	rsTopologyStatus := GenerateMultiClusterRootServiceStatus(rootService, rsCurrentStatus)

	rootService.Status.Topology = rsTopologyStatus
	return rootService
}
