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
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	"github.com/oceanbase/ob-operator/internal/telemetry/models"
)

type throttler struct {
	client     http.Client
	ctx        context.Context
	recordChan chan *models.TelemetryRecord
	filter     *expirable.LRU[string, struct{}]
}

var throttlerSingleton *throttler
var throttlerOnce sync.Once

func getThrottler(ctx context.Context) *throttler {
	throttlerOnce.Do(func() {
		cfg := obcfg.GetConfig().Telemetry
		throttlerSingleton = &throttler{
			recordChan: make(chan *models.TelemetryRecord, cfg.ThrottlerBufferSize),
		}

		throttlerSingleton.ctx = ctx
		throttlerSingleton.client = *http.DefaultClient
		throttlerSingleton.filter = expirable.NewLRU[string, struct{}](cfg.FilterSize, nil, cfg.FilterExpireTimeout)

		if u, err := url.Parse(cfg.Host); err == nil {
			clt := http.Client{
				Timeout: time.Second,
			}
			_, err := clt.Head(u.String())
			if err == nil {
				TelemetryReportScheme = u.Scheme
				TelemetryReportHost = u.Host
			}
		}

		throttlerSingleton.startWorkers()
		getLogger().Println("Telemetry throttler started", "#worker:", cfg.ThrottlerWorkerCount)
	})
	return throttlerSingleton
}

func (t *throttler) chanOut() <-chan *models.TelemetryRecord {
	return t.recordChan
}

func (t *throttler) chanIn() chan<- *models.TelemetryRecord {
	return t.recordChan
}

func (t *throttler) sendTelemetryRecord(record *models.TelemetryRecord) (*http.Response, error) {
	body, err := encodeRecord(record)
	if err != nil {
		return nil, err
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: TelemetryReportScheme,
			Host:   TelemetryReportHost,
			Path:   TelemetryReportPath,
		},
		Header: http.Header{
			"content-type": []string{ContentTypeJson},
		},
		Body: body,
	}
	return t.client.Do(req)
}

func (t *throttler) startWorkers() {
	for i := 0; i < obcfg.GetConfig().Telemetry.ThrottlerWorkerCount; i++ {
		go func(ctx context.Context, ch <-chan *models.TelemetryRecord) {
			for {
				select {
				case record, ok := <-ch:
					if !ok {
						// channel closed
						return
					}
					if uid, err := extractUID(record.Resource); err == nil {
						key := strings.Join([]string{record.ResourceType, uid, record.EventType, record.Reason, record.Message}, "-")
						if _, ok := t.filter.Get(key); ok {
							getLogger().Printf("Get the same key in filter: %s\n", key)
							continue
						}
						getLogger().Println("Add key to filter: ", key)
						getLogger().Println("Filter size: ", len(t.filter.Keys()))

						t.filter.Add(key, struct{}{})
					}
					res, err := t.sendTelemetryRecord(record)
					if err == nil && res != nil && res.Body != nil {
						if debugMode {
							getLogger().Printf("Report event: %v\n", record)
							bts, err := io.ReadAll(res.Body)
							if err != nil {
								getLogger().Printf("Read response body error: %v\n", err)
							} else {
								getLogger().Printf("[Event %s.%s] %s\n", record.ResourceType, record.EventType, string(bts))
							}
						} else {
							_, _ = io.Copy(io.Discard, res.Body)
						}
						_ = res.Body.Close()
					}
				case <-ctx.Done():
					getLogger().Println(ctx.Err())
					return
				}
			}
		}(t.ctx, t.chanOut())
	}
}
