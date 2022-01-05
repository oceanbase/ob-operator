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

	corev1 "k8s.io/api/core/v1"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	statefulapputil "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/util"
)

// pvc name: podname-pvcname
func GeneratePVCName(podName, volumeName string) string {
	pvcName := fmt.Sprintf("%s-%s", podName, volumeName)
	return pvcName
}

func GeneratePVCsObject(subsetName, podName string, podIndex int, statefulApp cloudv1.StatefulApp) []corev1.PersistentVolumeClaim {
	// TODO: support PVCSpecial
	res := make([]corev1.PersistentVolumeClaim, 0)
	pvcTemplates := statefulApp.Spec.StorageTemplates
	for _, pvcTemplate := range pvcTemplates {
		pvcName := GeneratePVCName(podName, pvcTemplate.Name)
		objectMeta := statefulapputil.GenerateObjectMeta(subsetName, pvcName, podIndex, statefulApp)
		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       pvcTemplate.PVC,
		}
		res = append(res, pvc)
	}
	return res
}

func PVCCurrentStatusToPVCStatus(pvc corev1.PersistentVolumeClaim) cloudv1.PVCStatus {
	var pvcStatus cloudv1.PVCStatus
	pvcStatus.Name = pvc.Name
	pvcStatus.Phase = corev1.PersistentVolumePhase(pvc.Status.Phase)
	return pvcStatus
}
