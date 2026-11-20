package common

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sEventTiming supports both legacy recorders and events.k8s.io recorders
// exposed through the core/v1 API, where the deprecated fields may be empty.
func K8sEventTiming(event corev1.Event) (first, last metav1.Time, count int32) {
	first = event.FirstTimestamp
	if !event.EventTime.IsZero() {
		first = metav1.NewTime(event.EventTime.Time)
	}
	if first.IsZero() {
		first = event.CreationTimestamp
	}
	last = event.LastTimestamp
	if event.Series != nil && !event.Series.LastObservedTime.IsZero() {
		last = metav1.NewTime(event.Series.LastObservedTime.Time)
	}
	if last.IsZero() {
		last = first
	}
	if first.IsZero() {
		first = last
	}
	count = event.Count
	if event.Series != nil && event.Series.Count > 0 {
		count = event.Series.Count
	}
	if count <= 0 {
		count = 1 // An existing, non-series event represents one occurrence.
	}
	return
}
