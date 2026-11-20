package response

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	corev1 "k8s.io/api/core/v1"
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
