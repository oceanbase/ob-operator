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

package oceanseek

import (
	"context"
	"fmt"
	"time"
)

const progressInterval = 30 * time.Second

// Keep quiet operations visible without exposing command output or arguments.
// The action retains its existing deadline and cancellation behavior. Wait for
// the reporter to stop before returning, so later phases cannot interleave.
func withProgress(ctx context.Context, interval time.Duration, progress Progress, label string, action func() error) error {
	started := time.Now()
	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				progress(fmt.Sprintf("[wait] %s (%s elapsed)", label, time.Since(started).Round(time.Second)))
			}
		}
	}()
	err := action()
	close(done)
	<-stopped
	return err
}
