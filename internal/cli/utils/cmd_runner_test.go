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
package utils_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

func TestAddHelmRepo(t *testing.T) {
	cmd := exec.Command("echo", "repo added")
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("failed to run command: %v", err)
	}

	err = utils.AddHelmRepo()
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestUpdateHelmRepo(t *testing.T) {
	cmd := exec.Command("echo", "repo updated")
	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("failed to run command: %v", err)
	}

	err = utils.UpdateHelmRepo()
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestBuildCmd(t *testing.T) {
	cmd, err := utils.BuildCmd("cert-manager", "2.2.2_release")
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	cmd, err = utils.BuildCmd("ob-operator", "2.3.0")
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	cmd, err = utils.BuildCmd("fake-component", "1.0.0")
	assert.Error(t, err)
	assert.Nil(t, cmd)
}

func TestRunCmd(t *testing.T) {
	cmd := exec.Command("echo", "running command")
	err := utils.RunCmd(cmd)
	assert.NoError(t, err)

	cmd = exec.Command("sh", "-c", "exit 1")
	err = utils.RunCmd(cmd)
	assert.Error(t, err)

	cmd, err = utils.BuildCmd("cert-manager", "2.2.2_release")
	assert.NoError(t, err)
	err = utils.RunCmd(cmd)
	assert.NoError(t, err)

}
