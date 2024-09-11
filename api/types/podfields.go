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

package types

import corev1 "k8s.io/api/core/v1"

type PodFieldsSpec struct {
	SchedulerName *string `json:"schedulerName,omitempty"`

	PriorityClassName *string                    `json:"priorityClassName,omitempty"`
	RuntimeClassName  *string                    `json:"runtimeClassName,omitempty"`
	PreemptionPolicy  *corev1.PreemptionPolicy   `json:"preemptionPolicy,omitempty"`
	Priority          *int32                     `json:"priority,omitempty"`
	SecurityContext   *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	DNSPolicy         *corev1.DNSPolicy          `json:"dnsPolicy,omitempty"`
	HostName          *string                    `json:"hostName,omitempty"`
	Subdomain         *string                    `json:"subdomain,omitempty"`

	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
}
