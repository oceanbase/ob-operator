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
	statefulappCore "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
)

func IsStatefulappInstanceReady(statefulappInstance cloudv1.StatefulApp) bool {
	if statefulappInstance.Status.ClusterStatus != statefulappCore.Ready {
		return false
	}
	for _, subset := range statefulappInstance.Status.Subsets {
		if subset.AvailableReplicas != subset.ExpectedReplicas {
			return false
		}
		for _, pod := range subset.Pods {
			if pod.PodPhase != statefulappCore.PodStatusRunning {
				return false
			}
			for _, pvc := range pod.PVCs {
				if pvc.Phase != statefulappCore.Bound {
					return false
				}
			}
		}
	}
	return true
}
