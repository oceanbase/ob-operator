package response

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewK8sEventTimestamps(t *testing.T) {
	first := time.Date(2026, 9, 20, 7, 0, 0, 0, time.UTC)
	last := first.Add(5 * time.Minute)
	for _, tc := range []struct {
		name        string
		event       corev1.Event
		first, last int64
		count       int32
	}{
		{"legacy", corev1.Event{FirstTimestamp: metav1.NewTime(first), LastTimestamp: metav1.NewTime(last), Count: 4}, first.Unix(), last.Unix(), 4},
		{"modern singleton", corev1.Event{EventTime: metav1.NewMicroTime(first)}, first.Unix(), first.Unix(), 1},
		{"modern series", corev1.Event{EventTime: metav1.NewMicroTime(first), Series: &corev1.EventSeries{Count: 7, LastObservedTime: metav1.NewMicroTime(last)}}, first.Unix(), last.Unix(), 7},
		{"creation fallback", corev1.Event{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(first)}}, first.Unix(), first.Unix(), 1},
		{"empty series", corev1.Event{EventTime: metav1.NewMicroTime(first), Series: &corev1.EventSeries{}}, first.Unix(), first.Unix(), 1},
		{"no timestamps", corev1.Event{}, 0, 0, 1},
		{"last timestamp only", corev1.Event{LastTimestamp: metav1.NewTime(last)}, last.Unix(), last.Unix(), 1},
		{"modern takes precedence", corev1.Event{EventTime: metav1.NewMicroTime(first), FirstTimestamp: metav1.NewTime(first.Add(-time.Hour)), LastTimestamp: metav1.NewTime(first), Count: 2, Series: &corev1.EventSeries{Count: 7, LastObservedTime: metav1.NewMicroTime(last)}}, first.Unix(), last.Unix(), 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.event.Namespace = "test"
			tc.event.Type = "Normal"
			tc.event.Reason = "Scheduled"
			tc.event.Message = "assigned"
			tc.event.InvolvedObject = corev1.ObjectReference{Kind: "Pod", Name: "test-pod"}
			got := NewK8sEvent(tc.event)
			if got.FirstOccur != tc.first || got.LastSeen != tc.last || got.Count != tc.count {
				t.Fatalf("timing = (%d, %d, %d), want (%d, %d, %d)", got.FirstOccur, got.LastSeen, got.Count, tc.first, tc.last, tc.count)
			}
			if got.Namespace != "test" || got.Object != "Pod/test-pod" || got.Type != "Normal" || got.Reason != "Scheduled" || got.Message != "assigned" {
				t.Fatalf("event identity changed: %+v", got)
			}
		})
	}
}
