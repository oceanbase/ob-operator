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

package tenantBackup

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func isRefController(controllerRef *metav1.OwnerReference) bool {
	refGV, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		klog.Errorf("could not parse OwnerReference %v APIVersion: %v", controllerRef, err)
		return false
	}
	return controllerRef.Kind == controllerKind.Kind && refGV.Group == controllerKind.Group
}

type tenantBackupEventHandler struct {
	enqueueHandler handler.EnqueueRequestForOwner
}

func (p *tenantBackupEventHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	tenantBackup := evt.Object.(*cloudv1.Backup)
	if tenantBackup.DeletionTimestamp != nil {
		p.Delete(event.DeleteEvent{Object: evt.Object}, q)
		return
	}
	controllerRef := metav1.GetControllerOf(tenantBackup)
	if controllerRef != nil && isRefController(controllerRef) {
		p.enqueueHandler.Create(evt, q)
	}
}

func (p *tenantBackupEventHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	oldBackup := evt.ObjectOld.(*cloudv1.Backup)
	newBackup := evt.ObjectNew.(*cloudv1.Backup)
	if newBackup.ResourceVersion == oldBackup.ResourceVersion {
		return
	}
	p.enqueueHandler.Update(evt, q)
}

func (p *tenantBackupEventHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	tenantBackup := evt.Object.(*cloudv1.Backup)
	controllerRef := metav1.GetControllerOf(tenantBackup)
	if controllerRef != nil && isRefController(controllerRef) {
		p.enqueueHandler.Delete(evt, q)
	}
}

func (p *tenantBackupEventHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}
