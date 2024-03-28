/*
Copyright (c) 2023 OceanBase
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func ListNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}

func CreateNamespace(ctx context.Context, namespace string) error {
	namespaceObject := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	client := client.GetClient()
	_, err := client.ClientSet.CoreV1().Namespaces().Create(ctx, &namespaceObject, metav1.CreateOptions{})
	return err
}
