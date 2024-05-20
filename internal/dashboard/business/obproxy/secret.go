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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func createPasswordSecret(ctx context.Context, ns, name, password string) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				constant.LabelManagedBy: constant.DASHBOARD_APP_NAME,
			},
		},
		StringData: map[string]string{
			"password": password,
		},
	}
	secret, err := client.GetClient().ClientSet.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to create secret, err msg: " + err.Error())
	}
	return secret, nil
}

// copyPasswordSecret copy a secret from another namespace to target namespace
func copyPasswordSecret(ctx context.Context, srcNs, srcName, tgtNs, tgtName string) (*corev1.Secret, error) {
	secret, err := client.GetClient().ClientSet.CoreV1().Secrets(srcNs).Get(ctx, srcName, metav1.GetOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to get secret, err msg: " + err.Error())
	}
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tgtName,
			Namespace: tgtNs,
			Labels: map[string]string{
				constant.LabelManagedBy: constant.DASHBOARD_APP_NAME,
			},
		},
		Data: secret.Data,
	}
	newSecret, err = client.GetClient().ClientSet.CoreV1().Secrets(tgtNs).Create(ctx, newSecret, metav1.CreateOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to create secret, err msg: " + err.Error())
	}
	return newSecret, nil
}
