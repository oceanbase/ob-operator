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
	"net/http"
	"net/url"
	"sync"
)

type throttler struct {
	http.Client

	metrics    *TelemetryEnvMetrics
	ctx        context.Context
	cancel     context.CancelFunc
	recordChan chan *TelemetryRecord
}

var throttlerSingleton *throttler
var throttlerOnce sync.Once

func getThrottler() *throttler {
	throttlerOnce.Do(func() {
		throttlerSingleton = &throttler{
			recordChan: make(chan *TelemetryRecord, DefaultThrottlerBufferSize),
		}
		if metrics, err := GetHostMetrics(); err == nil {
			throttlerSingleton.metrics = metrics
		}
		ctx, cancel := context.WithCancel(context.Background())
		throttlerSingleton.ctx = ctx
		throttlerSingleton.cancel = cancel
		throttlerSingleton.Client = *http.DefaultClient

		throttlerSingleton.startWorkers()
	})
	return throttlerSingleton
}

func (t *throttler) chanOut() <-chan *TelemetryRecord {
	return t.recordChan
}

func (t *throttler) chanIn() chan<- *TelemetryRecord {
	return t.recordChan
}

func (t *throttler) close() {
	if _, ok := <-t.recordChan; ok {
		close(t.recordChan)
	}
}

func (t *throttler) sendTelemetryRecord(record *TelemetryRecord) error {
	record.IpHashes = t.metrics.IPHashes
	body, err := encodeTelemetryRecord(record)
	req := &http.Request{
		Method: http.MethodPost,
		URL:    &url.URL{Host: TelemetryReportDevHost, Path: TelemetryReportPath},
		Header: http.Header{
			"content-type": []string{"application/json"},
			"sig":          []string{TelemetryRequestSignature},
		},
		Body: body,
	}
	_, err = t.Client.Do(req)
	return err
}

func (t *throttler) startWorkers() {
	for i := 0; i < DefaultWorkerCount; i++ {
		go func(ctx context.Context, ch <-chan *TelemetryRecord) error {
			for {
				select {
				case record, ok := <-ch:
					if !ok {
						// channel closed
						return nil
					}
					_ = t.sendTelemetryRecord(record)
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
			}
		}(t.ctx, t.chanOut())
	}
}
