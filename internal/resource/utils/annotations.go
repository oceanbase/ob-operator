/*
Copyright (c) 2023 OceanBase
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

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func GetCNIFromAnnotation(pod *corev1.Pod) string {
	_, found := pod.Annotations[oceanbaseconst.AnnotationCalicoValidate]
	if found {
		return oceanbaseconst.CNICalico
	}
	return oceanbaseconst.CNIUnknown
}

func NeedAnnotation(pod *corev1.Pod, cni string) bool {
	switch cni {
	case oceanbaseconst.CNICalico:
		_, found := pod.Annotations[oceanbaseconst.AnnotationCalicoIpAddrs]
		return !found
	default:
		return false
	}
}
