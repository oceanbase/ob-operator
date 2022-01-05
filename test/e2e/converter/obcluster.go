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
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
)

func IsOBClusterInstanceReady(obcluster cloudv1.OBCluster) bool {
	if obcluster.Status.Status != observerconst.TopologyReady {
		return false
	}
	for _, cluster := range obcluster.Status.Topology {
		if cluster.ClusterStatus != observerconst.ClusterReady {
			return false
		}
		for _, zone := range cluster.Zone {
			if zone.ZoneStatus != observerconst.OBZoneReady {
				return false
			}
			if zone.ExpectedReplicas != zone.AvailableReplicas {
				return false
			}
		}
	}
	return true
}
