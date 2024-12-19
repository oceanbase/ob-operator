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
package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// AddHelmRepo add ob-operator helm repo
func AddHelmRepo() error {
	repoURL := "https://oceanbase.github.io/ob-operator/"
	cmdAddRepo := exec.Command("helm", "repo", "add", "ob-operator", repoURL)
	if err := RunCmd(cmdAddRepo); err != nil {
		return fmt.Errorf("adding repo failed: %s", err)
	}
	return nil
}

// UpdateHelmRepo update ob-operator helm repo
func UpdateHelmRepo() error {
	cmdUpdateRepo := exec.Command("helm", "repo", "update")
	if err := RunCmd(cmdUpdateRepo); err != nil {
		return fmt.Errorf("updating repo failed: %s", err)
	}
	return nil
}

// RunCmd runs the command and prints the output to stdout
func RunCmd(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s command failed with error: %v", cmd.String(), err)
	}
	return nil
}

// BuildCmd builds the command based on the component and version
func BuildCmd(component, version string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	var (
		obURL         = "https://raw.githubusercontent.com/oceanbase/ob-operator/"
		localPathURL  = "https://raw.githubusercontent.com/rancher/local-path-provisioner/"
		dashboardRepo = "ob-operator/oceanbase-dashboard"
	)

	switch component {
	case "cert-manager":
		componentFile := "cert-manager.yaml"
		url := fmt.Sprintf("%s%s/deploy/%s", obURL, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-operator", "ob-operator-dev":
		componentFile := "operator.yaml"
		url := fmt.Sprintf("%s%s/deploy/%s", obURL, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "local-path-provisioner", "local-path-provisioner-dev":
		componentFile := "local-path-storage.yaml"
		url := fmt.Sprintf("%s%s/deploy/%s", localPathURL, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-dashboard":
		versionFlag := fmt.Sprintf("--version=%s", version)
		if err := AddHelmRepo(); err != nil {
			return nil, err
		}
		if err := UpdateHelmRepo(); err != nil {
			return nil, err
		}
		if !CheckIfComponentExists("ob-dashboard") {
			cmd = exec.Command("helm", "install", "oceanbase-dashboard", dashboardRepo, versionFlag)
		} else {
			cmd = exec.Command("helm", "upgrade", "oceanbase-dashboard", dashboardRepo, versionFlag)
		}
	default:
		return nil, fmt.Errorf("unknown component: %s", component)
	}
	return cmd, nil
}
