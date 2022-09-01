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
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	statefulappconst "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/judge"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

type SubsetCtrl struct {
	Resource    *resource.Resource
	StatefulApp cloudv1.StatefulApp
}

type SubsetCtrlOperator interface {
	SubsetCoordinator(ubset cloudv1.Subset) (bool, error)
	CreateSubset(subsetSpecName string, subsetsSpec []cloudv1.Subset) error
	GetSubsetsNameList() []string
	DeleteSubset(subsetSpecName string, subsetsCurrent []string) error
}

func NewSubsetCtrl(client client.Client, recorder record.EventRecorder, statefulApp cloudv1.StatefulApp) SubsetCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &SubsetCtrl{
		Resource:    ctrlResource,
		StatefulApp: statefulApp,
	}
}

func (ctrl *SubsetCtrl) SubsetCoordinator(subset cloudv1.Subset) (bool, error) {
	var compareStatus bool
	var err error

	podCtrl := NewPodCtrl(ctrl.Resource.Client, ctrl.Resource.Recorder, ctrl.StatefulApp)

	compareStatus = false

	subsetPodsCurrent := podCtrl.GetPodsBySubset(ctrl.StatefulApp.Namespace, ctrl.StatefulApp.Name, subset.Name)

	scaleState := judge.PodScale(int(subset.Replicas), len(subsetPodsCurrent))
	switch scaleState {
	case statefulappconst.ScaleUP:
		err = podCtrl.CreatePod(subset)
	case statefulappconst.ScaleDown:
		err = podCtrl.DeletePod(subset)
	case statefulappconst.Maintain:
		compareStatus, err = podCtrl.PodsCoordinator(subset, subsetPodsCurrent)
	}

	return compareStatus, err
}

func (ctrl *SubsetCtrl) CreateSubset(subsetSpecName string, subsetsSpec []cloudv1.Subset) error {
	var err error
	podCtrl := NewPodCtrl(ctrl.Resource.Client, ctrl.Resource.Recorder, ctrl.StatefulApp)
	for _, subsetSpec := range subsetsSpec {
		if subsetSpec.Name == subsetSpecName {
			// create one pod, then use Maintain to create more pods
			err = podCtrl.CreatePod(subsetSpec)
		}
	}
	return err
}

func (ctrl *SubsetCtrl) GetSubsetsNameList() []string {
	res := make([]string, 0)
	podCtrl := NewPodCtrl(ctrl.Resource.Client, ctrl.Resource.Recorder, ctrl.StatefulApp)
	pods := podCtrl.GetPodsByApp(ctrl.StatefulApp.Namespace, ctrl.StatefulApp.Name)
	if len(pods) == 0 {
		return res
	}
	subsetMap := converter.GetSubsetMapFromPods(pods)
	for k := range subsetMap {
		res = append(res, k)
	}
	return res
}

func (ctrl *SubsetCtrl) DeleteSubset(subsetSpecName string, subsetsCurrent []string) error {
	var err error
	klog.Infoln("DeleteSubset")
	podCtrl := NewPodCtrl(ctrl.Resource.Client, ctrl.Resource.Recorder, ctrl.StatefulApp)
	// zero
	if len(subsetsCurrent) == 1 {
		return errors.New("can't scale subsets to zero")
	}
	for _, subsetName := range subsetsCurrent {
		if subsetName == subsetSpecName {
			pods := podCtrl.GetPodsBySubset(ctrl.StatefulApp.Namespace, ctrl.StatefulApp.Name, subsetName)
			err = podCtrl.DeletePodList(pods)
		}
	}
	return err
}
