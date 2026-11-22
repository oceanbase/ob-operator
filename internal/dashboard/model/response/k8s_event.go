/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package response

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
)

func NewK8sEvent(event corev1.Event) K8sEvent {
	first, last, count := common.K8sEventTiming(event)
	result := K8sEvent{
		Namespace: event.Namespace, Type: event.Type, Count: count,
		Reason: event.Reason, Message: event.Message,
		Object: event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
	}
	// Zero is an unavailable timestamp, never the Unix value for year 0001.
	if !first.IsZero() {
		result.FirstOccur = first.Unix()
	}
	if !last.IsZero() {
		result.LastSeen = last.Unix()
	}
	return result
}
