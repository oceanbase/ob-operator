/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obproxy

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func getOBProxyService(ctx context.Context, ns, name string) (*corev1.Service, error) {
	svc, err := client.GetClient().ClientSet.CoreV1().Services(ns).Get(ctx, name+svcSuffix, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("Service not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy service, err msg: " + err.Error())
	}
	return svc, nil
}

func createOBProxyService(ctx context.Context, ns, name string, svcType corev1.ServiceType) (*corev1.Service, error) {
	svcName := name + svcSuffix
	svcParam := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: ns,
			Labels: map[string]string{
				LabelOBProxy:            name,
				constant.LabelManagedBy: constant.DASHBOARD_APP_NAME,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "sql",
				Port: 2883,
			}, {
				Name: "prometheus",
				Port: 2884,
			}},
			Selector: map[string]string{
				LabelOBProxy: name,
			},
			Type: svcType,
		},
	}
	svc, err := client.GetClient().ClientSet.CoreV1().Services(ns).Create(ctx, svcParam, metav1.CreateOptions{})
	if err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			return nil, httpErr.NewBadRequest("Service already exists")
		}
		return nil, httpErr.NewInternal("Failed to create obproxy service, err msg: " + err.Error())
	}
	return svc, nil
}

func updateOBProxyService(ctx context.Context, ns, name string, svcType corev1.ServiceType) (*corev1.Service, error) {
	svc, err := client.GetClient().ClientSet.CoreV1().Services(ns).Get(ctx, name+svcSuffix, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("Service not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy service, err msg: " + err.Error())
	}
	svc.Spec.Type = svcType
	svc, err = client.GetClient().ClientSet.CoreV1().Services(ns).Update(ctx, svc, metav1.UpdateOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to update obproxy service, err msg: " + err.Error())
	}
	return svc, nil
}

func deleteOBProxyService(ctx context.Context, ns, name string) (*corev1.Service, error) {
	svc, err := client.GetClient().ClientSet.CoreV1().Services(ns).Get(ctx, name+svcSuffix, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("Service not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy service, err msg: " + err.Error())
	}
	err = client.GetClient().ClientSet.CoreV1().Services(ns).Delete(ctx, name+svcSuffix, metav1.DeleteOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to delete obproxy service, err msg: " + err.Error())
	}
	return svc, nil
}
