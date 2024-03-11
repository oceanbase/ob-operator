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
	k8sconst "github.com/oceanbase/ob-operator/pkg/k8s/constants"
)

var timeout int64 = k8sconst.DefaultClientListTimeoutSeconds

func ListNodes() (*corev1.NodeList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}
