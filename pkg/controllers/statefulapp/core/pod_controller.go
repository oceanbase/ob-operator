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

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

type PodCtrl struct {
	Resource    *resource.Resource
	StatefulApp cloudv1.StatefulApp
}

type PodCtrlOperator interface {
	PodsCoordinator(subset cloudv1.Subset, subsetPodsCurrent []corev1.Pod) (bool, error)
	CreatePod(subset cloudv1.Subset) error
	GetPodsByLables(namespace string, listOption client.ListOption) []corev1.Pod
	GetPodsByApp(namespace, name string) []corev1.Pod
	GetPodsBySubset(namespace, name, subsetName string) []corev1.Pod
	GetPodsStatusBySubset(namespace, name, subsetName string) []cloudv1.PodStatus
	DeletePod(subset cloudv1.Subset) error
	DeletePodList(pods []corev1.Pod) error
}

func NewPodCtrl(client client.Client, recorder record.EventRecorder, statefulApp cloudv1.StatefulApp) PodCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &PodCtrl{
		Resource:    ctrlResource,
		StatefulApp: statefulApp,
	}
}

func (ctrl *PodCtrl) PodsCoordinator(subset cloudv1.Subset, subsetPodsCurrent []corev1.Pod) (bool, error) {
	var podNeedUpgrade bool
	podNeedUpgrade = false

	// index from large to small
	pods := converter.SortPodsDesc(subsetPodsCurrent)
	for _, pod := range pods {
		// TODO: support upgrade
		// compare spec to determine whether an upgrade is required
		if podNeedUpgrade {
			klog.Errorln("Not support upgrade pod yet", pod)
		}
	}

	return UpdateStatusWithEqual(ctrl.Resource, ctrl.StatefulApp)
}

func (ctrl *PodCtrl) CreatePod(subset cloudv1.Subset) error {
	var err error

	podsCurrent := ctrl.GetPodsBySubset(ctrl.StatefulApp.Namespace, ctrl.StatefulApp.Name, subset.Name)

	podName, podIndex, podObject := converter.GeneratePodObject(ctrl.StatefulApp, subset, podsCurrent)

	// create pod
	podExecuter := resource.NewPodResource(ctrl.Resource)
	err = podExecuter.Create(context.TODO(), podObject)
	if err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			klog.Errorln("pod is already exist, need recreate")
			// delete pod
			err = podExecuter.Delete(context.TODO(), podObject)
			if err != nil {
				return err
			}
		}
		return err
	}

	// create pvc
	pvcCtrl := NewPVCCtrl(ctrl.Resource.Client, ctrl.Resource.Recorder, ctrl.StatefulApp)
	err = pvcCtrl.CreatePVCs(subset, podName, podIndex)
	if err != nil {
		klog.Errorln("pvc is already exist, need recreate")
		// delete pvc
		err = pvcCtrl.DeletePVCs(podObject)
		if err != nil {
			return err
		}
		// delete pod
		err = podExecuter.Delete(context.TODO(), podObject)
		if err != nil {
			return err
		}
		return err
	}

	// update status
	if err == nil {
		err = UpdateStatus(ctrl.Resource, ctrl.StatefulApp)
		if err != nil {
			return err
		}
	}

	return err
}

func (ctrl *PodCtrl) GetPodsByLables(namespace string, listOption client.ListOption) []corev1.Pod {
	podExecuter := resource.NewPodResource(ctrl.Resource)
	podList := podExecuter.List(context.TODO(), namespace, listOption)
	return converter.PodListToPods(podList.(corev1.PodList))
}

func (ctrl *PodCtrl) GetPodsByApp(namespace, name string) []corev1.Pod {
	listOption := client.MatchingLabels{
		"app": name,
	}
	return ctrl.GetPodsByLables(namespace, listOption)
}

func (ctrl *PodCtrl) GetPodsBySubset(namespace, name, subsetName string) []corev1.Pod {
	listOption := client.MatchingLabels{
		"app":    name,
		"subset": subsetName,
	}
	return ctrl.GetPodsByLables(namespace, listOption)
}

func (ctrl *PodCtrl) GetPodsStatusBySubset(namespace, name, subsetName string) []cloudv1.PodStatus {
	res := make([]cloudv1.PodStatus, 0)
	podList := ctrl.GetPodsBySubset(namespace, name, subsetName)
	if len(podList) > 0 {
		for _, pod := range podList {
			podStatus := converter.PodCurrentStatusToPodStatus(pod)
			res = append(res, podStatus)
		}
	}
	return res
}

func (ctrl *PodCtrl) DeletePod(subset cloudv1.Subset) error {
	var err error
	var podIndex int
	var podName string

	podIndex = converter.GetDeleteIndex(ctrl.StatefulApp, subset)
	podName = converter.GeneratePodName(ctrl.StatefulApp.Name, myconfig.ClusterName, subset.Name, podIndex)
	pods := ctrl.GetPodsBySubset(ctrl.StatefulApp.Namespace, ctrl.StatefulApp.Name, subset.Name)

	// zero
	if len(pods) == 1 {
		return errors.New("can't scale pods to zero")
	}

	// delete pod
	podExecuter := resource.NewPodResource(ctrl.Resource)
	pvcCtrl := NewPVCCtrl(ctrl.Resource.Client, ctrl.Resource.Recorder, ctrl.StatefulApp)
	for _, pod := range pods {
		if pod.Name == podName {
			// TODO: support PVReclaimPolicy
			err = podExecuter.Delete(context.TODO(), pod)
			if err == nil {
				err = pvcCtrl.DeletePVCs(pod)
			}
			break
		}
	}

	// update status
	if err == nil {
		err = UpdateStatus(ctrl.Resource, ctrl.StatefulApp)
		if err != nil {
			return err
		}
	}

	return err
}

func (ctrl *PodCtrl) DeletePodList(pods []corev1.Pod) error {
	podExecuter := resource.NewPodResource(ctrl.Resource)
	var err error
	for _, pod := range pods {
		err = podExecuter.Delete(context.TODO(), pod)
	}
	return err
}
