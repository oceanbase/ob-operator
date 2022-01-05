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

package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddFinalizer(meta *metav1.ObjectMeta, finalizer string) {
	if !HasFinalizer(meta, finalizer) {
		meta.Finalizers = append(meta.Finalizers, finalizer)
	}
}

func HasFinalizer(meta *metav1.ObjectMeta, finalizer string) bool {
	return containsString(meta.Finalizers, finalizer)
}

func RemoveFinalizer(meta *metav1.ObjectMeta, finalizer string) {
	meta.Finalizers = removeString(meta.Finalizers, finalizer)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return result
}
