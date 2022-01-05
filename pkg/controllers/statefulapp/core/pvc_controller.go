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

package core

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

type PVCCtrl struct {
	Resource    *resource.Resource
	StatefulApp cloudv1.StatefulApp
}

type PVCCtrlOperator interface {
	CreatePVCs(subset cloudv1.Subset, podName string, podIndex int) error
	GetPVCsStatus(pod corev1.Pod) []cloudv1.PVCStatus
	DeletePVCs(pod corev1.Pod) error
}

func NewPVCCtrl(client client.Client, recorder record.EventRecorder, statefulApp cloudv1.StatefulApp) PVCCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &PVCCtrl{
		Resource:    ctrlResource,
		StatefulApp: statefulApp,
	}
}

func (ctrl *PVCCtrl) CreatePVCs(subset cloudv1.Subset, podName string, podIndex int) error {
	pvcs := converter.GeneratePVCsObject(subset.Name, podName, podIndex, ctrl.StatefulApp)
	pvcExecuter := resource.NewPVCResource(ctrl.Resource)
	err := pvcExecuter.Create(context.TODO(), pvcs)
	return err
}

func (ctrl *PVCCtrl) GetPVCsStatus(pod corev1.Pod) []cloudv1.PVCStatus {
	res := make([]cloudv1.PVCStatus, 0)
	pvcExecuter := resource.NewPVCResource(ctrl.Resource)
	for _, volume := range pod.Spec.Volumes {
		// pvc type
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.ClaimName
			pvcCurrent, err := pvcExecuter.Get(context.TODO(), pod.Namespace, pvcName)
			if err == nil {
				pvcStatus := converter.PVCCurrentStatusToPVCStatus(pvcCurrent.(corev1.PersistentVolumeClaim))
				res = append(res, pvcStatus)
			}
		}
	}
	return res
}

func (ctrl *PVCCtrl) DeletePVCs(pod corev1.Pod) error {
	var pvc corev1.PersistentVolumeClaim
	var pvcInterface interface{}
	var err error
	pvcExecuter := resource.NewPVCResource(ctrl.Resource)
	for _, volume := range pod.Spec.Volumes {
		// pvc type
		if volume.PersistentVolumeClaim != nil {
			pvcName := volume.PersistentVolumeClaim.ClaimName
			pvcInterface, err = pvcExecuter.Get(context.TODO(), pod.Namespace, pvcName)
			pvc = pvcInterface.(corev1.PersistentVolumeClaim)
			if err == nil {
				err = pvcExecuter.Delete(context.TODO(), pvc)
				if err != nil {
					break
				}
			}
		}
	}
	return err
}
