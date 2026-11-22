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
