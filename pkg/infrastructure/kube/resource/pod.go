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

package resource

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
)

type PodResource struct {
	*Resource
}

func NewPodResource(resource *Resource) ResourceOperator {
	return &PodResource{resource}
}

func (r *PodResource) Create(ctx context.Context, obj interface{}) error {
	pod := obj.(corev1.Pod)
	// kube.LogForAppActionStatus(pod.Kind, pod.Name, "create", pod)
	err := r.Client.Create(ctx, &pod)
	if err != nil {
		r.Recorder.Eventf(&pod, corev1.EventTypeWarning, FailedToCreatePod, "create Pod"+pod.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(pod.Kind, pod.Name, "create", "succeed")
	r.Recorder.Event(&pod, corev1.EventTypeNormal, CreatedPod, "create Pod"+pod.Name)
	return nil
}

func (r *PodResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	pod := &corev1.Pod{}
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), pod)
	if err != nil {
		klog.Errorln(err)
	}
	return *pod, err
}

func (r *PodResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	podList := &corev1.PodList{}
	err := r.Client.List(ctx, podList, client.InNamespace(namespace), listOption)
	if err != nil {
		// can definitely get a value, so errors are not returned
		klog.Errorln(err)
	}
	return *podList
}

func (r *PodResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *PodResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	var res error
	return res
}

func (r *PodResource) Delete(ctx context.Context, obj interface{}) error {
	pod := obj.(corev1.Pod)
	// kube.LogForAppActionStatus(pod.Kind, pod.Name, "delete", pod)
	err := r.Client.Delete(ctx, &pod)
	if err != nil {
		r.Recorder.Eventf(&pod, corev1.EventTypeWarning, FailedToKillPod, "delete Pod"+pod.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(pod.Kind, pod.Name, "delete", "succeed")
	r.Recorder.Event(&pod, corev1.EventTypeNormal, DeletedPod, "delete Pod"+pod.Name)
	return nil
}
