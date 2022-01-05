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

package util

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func GenerateOwnerReference(obCluster cloudv1.OBCluster) metav1.OwnerReference {
	ownerReference := metav1.OwnerReference{
		APIVersion: obCluster.APIVersion,
		Kind:       obCluster.Kind,
		Name:       obCluster.Name,
		UID:        obCluster.UID,
	}
	return ownerReference
}

func GenerateLabels(app string) map[string]string {
	labels := make(map[string]string)
	labels["app"] = app
	return labels
}

func GenerateObjectMeta(obCluster cloudv1.OBCluster, name string) metav1.ObjectMeta {
	ownerReference := GenerateOwnerReference(obCluster)
	labels := GenerateLabels(obCluster.Name)
	objectMeta := metav1.ObjectMeta{
		Name:            name,
		Namespace:       obCluster.Namespace,
		OwnerReferences: []metav1.OwnerReference{ownerReference},
		Labels:          labels,
	}
	return objectMeta
}
