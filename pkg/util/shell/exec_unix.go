//go:build !windows
// +build !windows

/*
Copyright (c) 2021 OceanBase
Copyright (c) 2015-2020 InfluxData Inc.
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package shell

import (
	"errors"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var TimeoutErr = errors.New("Command timed out.")

// KillGrace is the amount of time we allow a process to shutdown before sending a SIGKILL
const KillGrace = 5 * time.Second

// WaitTimeout waits for the given command to finish with a timeout
// It assumes the command has already been started
// If the command times out, it attempts to kill the process
func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	var kill *time.Timer
	term := time.AfterFunc(timeout, func() {
		err := c.Process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Errorf("[agent] Error terminating process: %s", err)
			return
		}
		kill = time.AfterFunc(KillGrace, func() {
			err := c.Process.Kill()
			if err != nil {
				log.Errorf("[agent] Error killing process: %s", err)
				return
			}
		})
	})

	err := c.Wait()

	// Shutdown all timers
	if kill != nil {
		kill.Stop()
	}
	termSent := !term.Stop()

	// If the process exited without error treat it as success
	// This allows a process to do a clean shutdown on signal
	if err == nil {
		return nil
	}

	// If SIGTERM was sent then treat any process error as a timeout
	if termSent {
		return TimeoutErr
	}

	// Otherwise there was an error unrelated to termination
	return err
}
