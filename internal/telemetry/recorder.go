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

package telemetry

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	record "k8s.io/client-go/tools/record"

	"github.com/oceanbase/ob-operator/internal/telemetry/models"
)

type Recorder interface {
	record.EventRecorder
	GenerateTelemetryRecord(object any, objectType, eventType, reason, message string, annotations map[string]string, extra ...models.ExtraField)
	GetHostMetrics() *hostMetrics
	Done()
}

type recorder struct {
	*throttler
	*hostMetrics
	record.EventRecorder

	ctx               context.Context
	telemetryDisabled bool
}

func NewRecorder(ctx context.Context, er record.EventRecorder) Recorder {
	clt := &recorder{
		ctx:           ctx,
		EventRecorder: er,
	}

	// if telemetry is disabled, return a dummy recorder as original event recorder
	if os.Getenv(DisableTelemetryEnvName) == "true" {
		clt.telemetryDisabled = true
		return clt
	}
	clt.hostMetrics = getHostMetrics()
	clt.throttler = getThrottler()
	return clt
}

// Implement record.EventRecorder interface
func (t *recorder) Event(object runtime.Object, eventType, reason, message string) {
	t.EventRecorder.Event(object, t.transformEventType(eventType), reason, message)
	t.generateFromEvent(object, nil, eventType, reason, message)
}

// Implement record.EventRecorder interface
func (t *recorder) Eventf(object runtime.Object, eventType, reason, messageFmt string, args ...any) {
	t.EventRecorder.Eventf(object, t.transformEventType(eventType), reason, messageFmt, args...)
	t.generateFromEvent(object, nil, eventType, reason, messageFmt, args...)
}

// Implement record.EventRecorder interface
func (t *recorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventType, reason, messageFmt string, args ...any) {
	t.EventRecorder.AnnotatedEventf(object, annotations, t.transformEventType(eventType), reason, messageFmt, args...)
	t.generateFromEvent(object, annotations, eventType, reason, messageFmt, args...)
}

// Use Event, Eventf, AnnotatedEventf first
func (t *recorder) GenerateTelemetryRecord(object any, objectType, eventType, reason, message string, annotations map[string]string, extra ...models.ExtraField) {
	if t.telemetryDisabled {
		return
	}
	go func(ctx context.Context, ch chan<- *models.TelemetryRecord) {
		objectSentry(object)
		record := newRecordFromEvent(object, objectType, eventType, reason, message, annotations, extra...)
		record.IpHashes = t.hostMetrics.IPHashes
		if object == nil && objectType == ObjectTypeOperator {
			record.Resource = t.hostMetrics
		}
		select {
		case ch <- record:
		case <-ctx.Done():
		default:
		}
	}(t.ctx, t.chanIn())
}

func (t *recorder) Done() {
	if t.telemetryDisabled {
		return
	}
	t.throttler.close()
}

func (t *recorder) GetHostMetrics() *hostMetrics {
	return t.hostMetrics
}

func (t *recorder) generateFromEvent(object runtime.Object, annotations map[string]string, eventType, reason, messageFmt string, args ...any) {
	if t.telemetryDisabled {
		return
	}
	if object == nil {
		t.GenerateTelemetryRecord(nil, ObjectTypeUnknown, eventType, reason, fmt.Sprintf(messageFmt, args...), annotations)
	} else {
		t.GenerateTelemetryRecord(object.DeepCopyObject(), object.GetObjectKind().GroupVersionKind().Kind, eventType, reason, fmt.Sprintf(messageFmt, args...), annotations)
	}
}

func (t *recorder) transformEventType(eventType string) string {
	// k8s EventRecorder only accepts `Warning` and `Normal` as event type
	switch eventType {
	case "Error", "Warning", "error", "warning":
		return corev1.EventTypeWarning
	default:
		return corev1.EventTypeNormal
	}
}
