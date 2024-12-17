/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package cli_test

import (
	"testing"

	"github.com/oceanbase/ob-operator/internal/cli"
)

func TestCli(t *testing.T) {
	// Test NewCliCmd
	cmd := cli.NewCliCmd()
	if cmd == nil {
		t.Errorf("NewCliCmd() failed")
	} else {
		t.Logf("NewCliCmd() success")
	}

	// Test Runable
	if err := cmd.RunE(cmd, []string{"help"}); err != nil {
		t.Errorf("RunE() failed")
	} else {
		t.Logf("RunE() success")
	}
}
