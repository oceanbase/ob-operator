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

package helper

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func TestGetLogServiceStartTimeout(t *testing.T) {
	defaultTimeout := time.Duration(oceanbaseconst.LogServiceStartTimeoutSeconds) * time.Second
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "default", value: "", want: defaultTimeout},
		{name: "configured", value: "42", want: 42 * time.Second},
		{name: "not an integer", value: "invalid", want: defaultTimeout},
		{name: "zero", value: "0", want: defaultTimeout},
		{name: "negative", value: "-1", want: defaultTimeout},
		{name: "overflow", value: "2147483648", want: defaultTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(logServiceStartTimeoutEnv, tt.value)
			if got := getLogServiceStartTimeout(); got != tt.want {
				t.Fatalf("getLogServiceStartTimeout() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestStartAndWaitForLauncherReturnsExitError(t *testing.T) {
	_, err := startAndWaitForLauncher(exec.Command("/bin/sh", "-c", "exit 7"), time.Second)
	if err == nil {
		t.Fatal("startAndWaitForLauncher() returned nil, want a launcher exit error")
	}
	if !strings.Contains(err.Error(), "launcher exited with an error") {
		t.Fatalf("startAndWaitForLauncher() error = %q, want launcher exit context", err)
	}
}

func TestStartAndWaitForLauncherKillsTimedOutProcess(t *testing.T) {
	cmd := exec.Command("/bin/sleep", "10")
	startedAt := time.Now()
	_, err := startAndWaitForLauncher(cmd, 20*time.Millisecond)
	if err == nil {
		t.Fatal("startAndWaitForLauncher() returned nil, want a timeout error")
	}
	if !strings.Contains(err.Error(), "failed to exit within 20ms") {
		t.Fatalf("startAndWaitForLauncher() error = %q, want timeout context", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("startAndWaitForLauncher() took %s, want it to terminate the launcher promptly", elapsed)
	}
	if cmd.ProcessState == nil {
		t.Fatal("startAndWaitForLauncher() did not reap the timed-out launcher")
	}
}
