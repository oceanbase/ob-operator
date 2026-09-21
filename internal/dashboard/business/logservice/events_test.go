package logservice

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDetailModernEventTime(t *testing.T) {
	s, _ := fixture(t)
	ctx := context.Background()
	first := time.Date(2026, 9, 20, 7, 0, 0, 0, time.UTC)
	last := first.Add(time.Minute)
	for _, e := range []corev1.Event{
		{ObjectMeta: metav1.ObjectMeta{Name: "older", Namespace: "test"}, EventTime: metav1.NewMicroTime(first)},
		{ObjectMeta: metav1.ObjectMeta{Name: "newer", Namespace: "test"}, EventTime: metav1.NewMicroTime(first), Series: &corev1.EventSeries{Count: 2, LastObservedTime: metav1.NewMicroTime(last)}},
	} {
		e.InvolvedObject = corev1.ObjectReference{Name: "ls-test", Kind: "OBLogServiceCluster"}
		e.Reason = e.Name
		if _, err := s.Core.CoreV1().Events("test").Create(ctx, &e, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	detail, err := s.Get(ctx, "test", "ls-test")
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Events) != 2 || detail.Events[0].Reason != "newer" || !detail.Events[0].Time.Time.Equal(last) || !detail.Events[1].Time.Time.Equal(first) {
		t.Fatalf("unexpected event ordering/timestamps: %+v", detail.Events)
	}
}
