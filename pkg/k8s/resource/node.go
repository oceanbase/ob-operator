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
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	k8sconst "github.com/oceanbase/ob-operator/pkg/k8s/constants"
)

var timeout int64 = k8sconst.DefaultClientListTimeoutSeconds

func UpdateNode(ctx context.Context, node *corev1.Node) (*corev1.Node, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
}

func GetNode(ctx context.Context, name string) (*corev1.Node, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

func ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}

func ListNodeMetrics(ctx context.Context) (map[string]metricsv1beta1.NodeMetrics, error) {
	client := client.GetClient()
	nodeMetricsMap := make(map[string]metricsv1beta1.NodeMetrics)
	metricsList, err := client.MetricsClientset.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
	if err == nil {
		for _, metrics := range metricsList.Items {
			nodeMetricsMap[metrics.Name] = metrics
		}
	}
	return nodeMetricsMap, err
}
