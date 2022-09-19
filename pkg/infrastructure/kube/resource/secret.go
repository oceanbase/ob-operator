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

type SecretResource struct {
	*Resource
}

func NewSecretResource(resource *Resource) ResourceOperator {
	return &SecretResource{resource}
}

func (r *SecretResource) Create(ctx context.Context, obj interface{}) error {
	secret := obj.(corev1.Secret)
	// kube.LogForAppActionStatus(service.Kind, service.Name, "create", service)
	err := r.Client.Create(ctx, &secret)
	if err != nil {
		r.Recorder.Eventf(&secret, corev1.EventTypeWarning, FailedToCreateSecret, "create Secret"+secret.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(secret.Kind, secret.Name, "create", "succeed")
	r.Recorder.Event(&secret, corev1.EventTypeNormal, CreatedSecret, "create Secret"+secret.Name)
	return nil
}

func (r *SecretResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), secret)
	if err != nil {
		klog.Errorln(err)
	}
	return *secret, err
}

func (r *SecretResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	return nil
}

func (r *SecretResource) Update(ctx context.Context, obj interface{}) error {
	secret := obj.(corev1.Secret)
	err := r.Client.Update(ctx, &secret)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *SecretResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	var res error
	return res
}

func (r *SecretResource) Delete(ctx context.Context, obj interface{}) error {
	secret := obj.(corev1.Secret)
	// kube.LogForAppActionStatus(service.Kind, service.Name, "delete", service)
	err := r.Client.Delete(ctx, &secret)
	if err != nil {
		r.Recorder.Eventf(&secret, corev1.EventTypeWarning, FailedToDeleteSecret, "delete secret"+secret.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(secret.Kind, secret.Name, "delete", "succeed")
	r.Recorder.Event(&secret, corev1.EventTypeNormal, DeletedSecret, "delete secret"+secret.Name)
	return nil
}
