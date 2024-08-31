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
package install

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	utils "github.com/oceanbase/ob-operator/internal/cli/utils"
)

type InstallOptions struct {
	version    string
	Components map[string]string
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{
		Components: utils.GetComponentsConf(),
	}
}

func (o *InstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.version, "version", "", "version of component")
}

func (o *InstallOptions) Parse(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	name := args[0]
	if v, ok := o.Components[name]; ok {
		if o.version != "" {
			o.Components = map[string]string{name: o.version}
		} else {
			o.Components = map[string]string{name: v}
		}
		return nil
	}
	return fmt.Errorf("%s install not supported", name)
}

// Install component
func Install(component string, version string) error {
	var url string
	var cmd *exec.Cmd
	obUrl := "https://raw.githubusercontent.com/oceanbase/ob-operator/"
	localPathUrl := "https://raw.githubusercontent.com/rancher/local-path-provisioner/"
	switch component {
	case "cert-manager":
		componentFile := "cert-manager.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", obUrl, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-operator", "ob-operator-dev":
		componentFile := "operator.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", obUrl, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "local-path-provisioner", "local-path-provisioner-dev":
		componentFile := "local-path-storage.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", localPathUrl, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-dashboard":
		if err := preRunCmd(); err != nil {
			return err
		}
		versionFlag := fmt.Sprintf("--version=%s", version)
		cmd = exec.Command("helm", "install", "oceanbase-dashboard", "ob-operator/oceanbase-dashboard", versionFlag)
	}
	if err := runCmd(cmd); err != nil {
		return err
	}
	return nil
}

// preRunCmd preRun two commands for ob-dashboard installation
func preRunCmd() error {
	// add helm repo
	cmdAddRepo := exec.Command("helm", "repo", "add", "ob-operator", "https://oceanbase.github.io/ob-operator/")
	output, err := cmdAddRepo.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adding repo failed: %s, %s", err, output)
	}

	// update helm repo
	cmdUpdateRepo := exec.Command("helm", "repo", "update", "ob-operator")
	output, err = cmdUpdateRepo.CombinedOutput()
	if err != nil {
		return fmt.Errorf("updating repo failed: %s, %s", err, output)
	}

	return nil
}

func runCmd(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed with error: %v", err)
	}
	return nil
}
