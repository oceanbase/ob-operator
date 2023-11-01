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

	"k8s.io/apimachinery/pkg/runtime"
	record "k8s.io/client-go/tools/record"
)

type Telemetry interface {
	record.EventRecorder
	GenerateTelemetryRecord(object any, objectType, eventType, reason, message string, annotations map[string]string, extra ...ExtraField)
}

type telemetry struct {
	*throttler
	record.EventRecorder

	telemetryDisabled bool
}

// Implement record.EventRecorder interface
func (t *telemetry) Event(object runtime.Object, eventType, reason, message string) {
	t.EventRecorder.Event(object, eventType, reason, message)
	t.GenerateTelemetryRecord(object.DeepCopyObject(), object.GetObjectKind().GroupVersionKind().Kind, eventType, reason, message, nil)
}

// Implement record.EventRecorder interface
func (t *telemetry) Eventf(object runtime.Object, eventType, reason, messageFmt string, args ...interface{}) {
	t.EventRecorder.Eventf(object, eventType, reason, messageFmt, args...)
	t.GenerateTelemetryRecord(object.DeepCopyObject(), object.GetObjectKind().GroupVersionKind().Kind, eventType, reason, fmt.Sprintf(messageFmt, args...), nil)
}

// Implement record.EventRecorder interface
func (t *telemetry) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventType, reason, messageFmt string, args ...interface{}) {
	t.EventRecorder.AnnotatedEventf(object, annotations, eventType, reason, messageFmt, args...)
	t.GenerateTelemetryRecord(object.DeepCopyObject(), object.GetObjectKind().GroupVersionKind().Kind, eventType, reason, fmt.Sprintf(messageFmt, args...), annotations)
}

func (t *telemetry) GenerateTelemetryRecord(object any, objectType, eventType, reason, message string, annotations map[string]string, extra ...ExtraField) {
	if t.telemetryDisabled {
		return
	}
	go func(ctx context.Context, ch chan<- *TelemetryRecord) {
		// TODO: guard here to mask IP address
		select {
		case <-ctx.Done():
			return
		case ch <- newTelemetryRecord(object, objectType, eventType, reason, message, annotations, extra...):
		default:
		}
	}(t.ctx, t.chanIn())
}

func NewTelemetry(recorder record.EventRecorder) Telemetry {
	clt := &telemetry{
		EventRecorder: recorder,
	}

	if disabled, exist := os.LookupEnv(DisableTelemetryEnvName); exist && disabled == "true" {
		clt.telemetryDisabled = true
		return clt
	}

	clt.throttler = getThrottler()
	return clt
}

func (t *telemetry) Done() {
	t.cancel()
}
