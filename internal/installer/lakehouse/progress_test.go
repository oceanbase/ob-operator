package lakehouse

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestProgressHeartbeatPreservesResultAndStops(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	observed := make(chan string, 100)
	want := errors.New("private command failure")
	err := withProgress(ctx, time.Millisecond, func(line string) {
		observed <- line
	}, "Importing core images", func() error {
		select {
		case line := <-observed:
			if !strings.Contains(line, "[wait] Importing core images") || !strings.Contains(line, "elapsed") {
				t.Errorf("missing useful heartbeat: %s", line)
			}
		case <-ctx.Done():
			t.Error("quiet operation did not report progress")
		}
		return want
	})
	if err != want {
		t.Fatalf("action failure was replaced: %v", err)
	}
	// The function joins its reporter; a following phase must never receive a
	// late heartbeat from the previous action.
	for len(observed) > 0 {
		<-observed
	}
	select {
	case line := <-observed:
		t.Fatalf("reporter outlived the action: %s", line)
	case <-time.After(5 * time.Millisecond):
	}
}

func TestFastActionDoesNotPrintWait(t *testing.T) {
	err := withProgress(context.Background(), time.Hour, func(line string) {
		t.Errorf("fast action printed a wait: %s", line)
	}, "Waiting for Operator", func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
}
