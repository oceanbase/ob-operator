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

package judge

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
)

func OBServerScale(clusterList []cloudv1.Cluster, statefulApp cloudv1.StatefulApp) (string, cloudv1.Subset) {
	var res cloudv1.Subset
	var scaleState string
	cluster := converter.GetClusterSpecFromOBTopology(clusterList)
	for _, zone := range cluster.Zone {
		subsetStatus := converter.GetSubsetStatusFromStatefulApp(zone.Name, statefulApp)
		if zone.Replicas > subsetStatus.ExpectedReplicas {
			scaleState = observerconst.ScaleUP
			res = zone
			break
		} else if zone.Replicas < subsetStatus.ExpectedReplicas {
			scaleState = observerconst.ScaleDown
			res = zone
			break
		} else {
			scaleState = observerconst.Maintain
		}
	}
	return scaleState, res
}
