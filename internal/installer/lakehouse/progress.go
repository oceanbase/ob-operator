package lakehouse

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
