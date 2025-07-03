/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package monitor

import (
	"context"
	"fmt"

	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

func ListEndpoints(ctx context.Context) ([]response.MonitorEndpoint, error) {
	endpoints := make([]response.MonitorEndpoint, 0)
	observerList := v1alpha1.OBServerList{}
	err := clients.ServerClient.List(ctx, corev1.NamespaceAll, &observerList, metav1.ListOptions{})
	if err != nil {
		logger.WithError(err).Error("Failed to list all observers")
		return endpoints, err
	}
	targets := make([]string, 0)
	for _, observer := range observerList.Items {
		if observer.Spec.MonitorTemplate != nil {
			targets = append(targets, fmt.Sprintf("%s:%d", observer.Status.PodIp, obagentconst.HttpPort))
		}
	}
	endpoint := response.MonitorEndpoint{
		Targets: targets,
	}
	endpoints = append(endpoints, endpoint)
	return endpoints, nil
}
