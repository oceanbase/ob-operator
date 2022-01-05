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
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func GenerateOwnerReference(statefulApp cloudv1.StatefulApp) metav1.OwnerReference {
	ownerReference := metav1.OwnerReference{
		APIVersion: statefulApp.APIVersion,
		Kind:       statefulApp.Kind,
		Name:       statefulApp.Name,
		UID:        statefulApp.UID,
	}
	return ownerReference
}

func GenerateLabels(app, subset, index string) map[string]string {
	labels := make(map[string]string)
	labels["app"] = app
	labels["subset"] = subset
	labels["index"] = index
	return labels
}

func GenerateObjectMeta(subsetName string, name string, index int, statefulApp cloudv1.StatefulApp) metav1.ObjectMeta {
	ownerReference := GenerateOwnerReference(statefulApp)
	labels := GenerateLabels(statefulApp.Name, subsetName, strconv.Itoa(index))
	objectMeta := metav1.ObjectMeta{
		Name:            name,
		Namespace:       statefulApp.Namespace,
		OwnerReferences: []metav1.OwnerReference{ownerReference},
		Labels:          labels,
	}
	return objectMeta
}
