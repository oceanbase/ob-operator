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

package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func GetSQLAnalyzerPodIP(ctx context.Context, namespace, obtenant string) (string, error) {
	k8sclient := client.GetClient()
	pods, err := k8sclient.ClientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=sql-analyzer,tenant=%s", obtenant),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to list sql-analyzer pods")
	}
	if len(pods.Items) == 0 {
		return "", errors.New("no sql-analyzer pod found")
	}
	return pods.Items[0].Status.PodIP, nil
}
