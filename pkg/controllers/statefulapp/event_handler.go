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

package statefulapp

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

func isRefController(controllerRef *metav1.OwnerReference) bool {
	refGV, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		klog.Errorf("could not parse OwnerReference %v APIVersion: %v", controllerRef, err)
		return false
	}
	return controllerRef.Kind == controllerKind.Kind && refGV.Group == controllerKind.Group
}

type podEventHandler struct {
	enqueueHandler handler.EnqueueRequestForOwner
}

func (p *podEventHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	pod := evt.Object.(*v1.Pod)
	if pod.DeletionTimestamp != nil {
		p.Delete(event.DeleteEvent{Object: evt.Object}, q)
		return
	}
	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef != nil && isRefController(controllerRef) {
		p.enqueueHandler.Create(evt, q)
	}
}

func (p *podEventHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	oldPod := evt.ObjectOld.(*v1.Pod)
	newPod := evt.ObjectNew.(*v1.Pod)
	if newPod.ResourceVersion == oldPod.ResourceVersion {
		return
	}
	p.enqueueHandler.Update(evt, q)
}

func (p *podEventHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	pod := evt.Object.(*v1.Pod)
	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef != nil && isRefController(controllerRef) {
		p.enqueueHandler.Delete(evt, q)
	}
}

func (p *podEventHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}

type pvcEventHandler struct {
	enqueueHandler handler.EnqueueRequestForOwner
}

func (p *pvcEventHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	pvc := evt.Object.(*v1.PersistentVolumeClaim)
	if pvc.DeletionTimestamp != nil {
		p.Delete(event.DeleteEvent{Object: evt.Object}, q)
		return
	}
	controllerRef := metav1.GetControllerOf(pvc)
	if controllerRef != nil && isRefController(controllerRef) {
		p.enqueueHandler.Create(evt, q)
	}
}

func (p *pvcEventHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	oldPVC := evt.ObjectOld.(*v1.PersistentVolumeClaim)
	newPVC := evt.ObjectNew.(*v1.PersistentVolumeClaim)
	if newPVC.ResourceVersion == oldPVC.ResourceVersion {
		return
	}
	p.enqueueHandler.Update(evt, q)
}

func (p *pvcEventHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	pvc := evt.Object.(*v1.PersistentVolumeClaim)
	controllerRef := metav1.GetControllerOf(pvc)
	if controllerRef != nil && isRefController(controllerRef) {
		p.enqueueHandler.Delete(evt, q)
	}
}

func (p *pvcEventHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}
