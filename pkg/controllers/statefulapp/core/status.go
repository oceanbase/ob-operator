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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	statefulappconst "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func UpdateStatus(ctrlResource *resource.Resource, statefulApp cloudv1.StatefulApp) error {
	statefulAppExecuter := resource.NewStatefulAppResource(ctrlResource)
	// use retry to update
	retryErr := retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {
			// get current StatefulApp every time
			statefulAppTemp, err := statefulAppExecuter.Get(context.TODO(), statefulApp.Namespace, statefulApp.Name)
			if err != nil {
				return err
			}
			// DeepCopy
			statefulAppCurrent := statefulAppTemp.(cloudv1.StatefulApp)
			statefulAppCurrentDeepCopy := statefulAppCurrent.DeepCopy()
			// build status
			statefulAppNew := buildStatus(ctrlResource, *statefulAppCurrentDeepCopy)
			// update status
			err = statefulAppExecuter.UpdateStatus(context.TODO(), statefulAppNew)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if retryErr != nil {
		klog.Errorln(retryErr)
		return retryErr
	}
	kube.LogForAppActionStatus(statefulApp.Kind, statefulApp.Name, "update status", statefulApp.Status)
	return nil
}

func UpdateStatusWithEqual(ctrlResource *resource.Resource, statefulApp cloudv1.StatefulApp) (bool, error) {
	statefulAppExecuter := resource.NewStatefulAppResource(ctrlResource)
	// get current
	statefulAppTemp, err := statefulAppExecuter.Get(context.TODO(), statefulApp.Namespace, statefulApp.Name)
	if err != nil {
		// this one is broken, try next
		return true, err
	}
	// DeepCopy
	statefulAppCurrent := statefulAppTemp.(cloudv1.StatefulApp)
	statefulAppCurrentDeepCopy := statefulAppCurrent.DeepCopy()
	// build status
	statefulAppNew := buildStatus(ctrlResource, *statefulAppCurrentDeepCopy)
	// compare status, if Equal don't need update
	compareStatus := reflect.DeepEqual(statefulAppCurrent.Status, statefulAppNew.Status)
	if compareStatus {
		return compareStatus, nil
	}
	// update status
	err = statefulAppExecuter.UpdateStatus(context.TODO(), statefulAppNew)
	if err != nil {
		klog.Errorln(err)
		// this one is broken, try next
		return true, err
	}
	kube.LogForAppActionStatus(statefulApp.Kind, statefulApp.Name, "update status", statefulApp.Status)
	// do one thing at a time
	return false, nil
}

// update all status every time
func buildStatus(ctrlResource *resource.Resource, statefulApp cloudv1.StatefulApp) cloudv1.StatefulApp {
	subsetsStatus := make([]cloudv1.SubsetStatus, 0)
	podCtrl := NewPodCtrl(ctrlResource.Client, ctrlResource.Recorder, statefulApp)
	pvcCtrl := NewPVCCtrl(ctrlResource.Client, ctrlResource.Recorder, statefulApp)
	for _, subset := range statefulApp.Spec.Subsets {
		subsetStatus := buildSubsetStatus(podCtrl, pvcCtrl, statefulApp, subset)
		subsetsStatus = append(subsetsStatus, subsetStatus)
	}
	// check cluster status
	clusterStatus := checkClusterStatus(statefulApp)
	statefulApp = generateStatefulAppStatus(statefulApp.Spec.Cluster, clusterStatus, subsetsStatus, statefulApp)
	return statefulApp
}

func buildSubsetStatus(podCtrl PodCtrlOperator, pvcCtrl PVCCtrlOperator, statefulApp cloudv1.StatefulApp, subset cloudv1.Subset) cloudv1.SubsetStatus {
	// pod status
	var replicas int
	podsStatus := make([]cloudv1.PodStatus, 0)
	pods := podCtrl.GetPodsBySubset(statefulApp.Namespace, statefulApp.Name, subset.Name)
	if len(pods) > 0 {
		for _, pod := range pods {
			podStatus := buildPodStatus(pvcCtrl, pod)
			podsStatus = append(podsStatus, podStatus)
			if podStatus.PodPhase == statefulappconst.PodStatusRunning {
				replicas += 1
			}
		}
	}
	podsStatus = converter.SortPodsStatus(podsStatus)
	subsetStatus := converter.GenerateSubsetStatus(subset.Name, subset.Region, subset.Replicas, int32(replicas), podsStatus)
	return subsetStatus
}

func buildPodStatus(pvcCtrl PVCCtrlOperator, pod corev1.Pod) cloudv1.PodStatus {
	podStatus := converter.PodCurrentStatusToPodStatus(pod)
	pvcs := pvcCtrl.GetPVCsStatus(pod)
	if len(pvcs) > 0 {
		podStatus.PVCs = pvcs
	}
	return podStatus
}

func checkClusterStatus(statefulApp cloudv1.StatefulApp) string {
	var status string
	status = statefulappconst.Ready
	// subset number is right
	if len(statefulApp.Status.Subsets) == len(statefulApp.Spec.Subsets) {
		for _, subset := range statefulApp.Status.Subsets {
			// replica number is right
			if subset.ExpectedReplicas != subset.AvailableReplicas {
				status = statefulappconst.Prepareing
				break
			}
			// pod status is Running
			for _, pod := range subset.Pods {
				if pod.PodPhase != statefulappconst.PodStatusRunning {
					status = statefulappconst.Prepareing
					break
				}
				// pvc status is Bound
				for _, pv := range pod.PVCs {
					if pv.Phase != statefulappconst.Bound {
						status = statefulappconst.Prepareing
						break
					}
				}
			}
		}
	}
	return status
}

func generateStatefulAppStatus(clusterName, clusterStatus string, topology []cloudv1.SubsetStatus, statefulApp cloudv1.StatefulApp) cloudv1.StatefulApp {
	statefulApp.Status.Cluster = clusterName
	statefulApp.Status.ClusterStatus = clusterStatus
	statefulApp.Status.Subsets = topology
	return statefulApp
}
