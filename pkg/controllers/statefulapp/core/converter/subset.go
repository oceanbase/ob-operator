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
	corev1 "k8s.io/api/core/v1"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func GenerateSubsetStatus(subsetName, regionName string, expectedReplicas, availableReplicas int32, podsStatus []cloudv1.PodStatus) cloudv1.SubsetStatus {
	var subsetStatus cloudv1.SubsetStatus
	subsetStatus.Name = subsetName
	if regionName != "" {
		subsetStatus.Region = regionName
	}
	subsetStatus.ExpectedReplicas = expectedReplicas
	subsetStatus.AvailableReplicas = availableReplicas
	subsetStatus.Pods = podsStatus
	return subsetStatus
}

func FindElementNotInSubsetsCurrentNameList(subsetsSpec []cloudv1.Subset, subsetsCurrentNameList []string) string {
	var res string
	for _, subsetSpec := range subsetsSpec {
		var matchStatus bool
		matchStatus = false
		for _, subsetName := range subsetsCurrentNameList {
			if subsetSpec.Name == subsetName {
				matchStatus = true
			}
		}
		if !matchStatus {
			res = subsetSpec.Name
			break
		}
	}
	return res
}

func FindElementNotInSubsetsSpec(subsetsSpec []cloudv1.Subset, subsetsCurrentNameList []string) string {
	var res string
	for _, subsetName := range subsetsCurrentNameList {
		var matchStatus bool
		matchStatus = false
		for _, subsetSpec := range subsetsSpec {
			if subsetSpec.Name == subsetName {
				matchStatus = true
			}
		}
		if !matchStatus {
			res = subsetName
			break
		}
	}
	return res
}

func GetSubsetMapFromPods(pods []corev1.Pod) map[string]bool {
	subsetMap := make(map[string]bool)
	for _, pod := range pods {
		key := pod.Labels["subset"]
		subsetMap[key] = true
	}
	return subsetMap
}
