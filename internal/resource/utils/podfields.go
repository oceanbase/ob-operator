/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package utils

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/oceanbase/ob-operator/api/types"
)

// GetSchedulerName returns the scheduler name from the given PodFields.
// If PodFields or SchedulerName is nil, it returns the default scheduler name.
func GetSchedulerName(podFields *types.PodFieldsSpec) string {
	if podFields == nil || podFields.SchedulerName == nil {
		return corev1.DefaultSchedulerName // return Kubernetes's default scheduler name
	}
	return *podFields.SchedulerName
}
