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
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackupResource struct {
	*Resource
}

func NewBackupResource(resource *Resource) ResourceOperator {
	return &BackupResource{resource}
}

func (r *BackupResource) Create(ctx context.Context, obj interface{}) error {
	backup := obj.(cloudv1.Backup)
	err := r.Client.Create(ctx, &backup)
	if err != nil {
		r.Recorder.Eventf(&backup, corev1.EventTypeWarning, FailedToCreateBackup, "create Backup"+backup.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(backup.Kind, backup.Name, "create", "succeed")
	r.Recorder.Event(&backup, corev1.EventTypeNormal, CreatedBackup, "create Backup"+backup.Name)
	return nil
}

func (r *BackupResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var backupCurrent cloudv1.Backup
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &backupCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return backupCurrent, err
}

func (r *BackupResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	backupList := &cloudv1.BackupList{}
	err := r.Client.List(ctx, backupList, client.InNamespace(namespace), listOption)
	if err != nil {
		// can definitely get a value, so errors are not returned
		klog.Errorln(err)
	}
	return *backupList
}

func (r *BackupResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *BackupResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	backup := obj.(cloudv1.Backup)
	err := r.Client.Status().Update(ctx, &backup)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *BackupResource) Delete(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *BackupResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
