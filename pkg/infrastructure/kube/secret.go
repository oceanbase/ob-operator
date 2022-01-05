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

package kube

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LoadSecretData(apiReader client.Reader, namespace, secretName, dataKey string) (string, error) {
	secret := &corev1.Secret{}
	err := apiReader.Get(context.TODO(), GenerateNamespacedName(namespace, secretName), secret)
	if err != nil {
		return "", err
	}
	retStr, ok := secret.Data[dataKey]
	if !ok {
		return "", errors.Errorf("secret %s did not contain key %s", secretName, dataKey)
	}
	return string(retStr), nil
}

func LoadSecretDataUsingClient(c client.Client, namespace, secretName, dataKey string) (string, error) {
	secret := &corev1.Secret{}
	err := c.Get(context.TODO(), GenerateNamespacedName(namespace, secretName), secret)
	if err != nil {
		return "", err
	}
	retStr, ok := secret.Data[dataKey]
	if !ok {
		return "", errors.Errorf("secret %s did not contain key %s", secretName, dataKey)
	}
	return string(retStr), nil
}
