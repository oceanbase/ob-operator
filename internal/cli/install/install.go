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
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	utils "github.com/oceanbase/ob-operator/internal/cli/utils"
)

type InstallOptions struct {
	version      string
	Components   map[string]string
	obUrl        string
	localPathUrl string
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{
		Components:   utils.GetComponentsConf(),
		obUrl:        "https://raw.githubusercontent.com/oceanbase/ob-operator/",
		localPathUrl: "https://raw.githubusercontent.com/rancher/local-path-provisioner/",
	}
}

func (o *InstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.version, "version", "", "version of component")
}

func (o *InstallOptions) Parse(_ *cobra.Command, args []string) error {
	// if not specified, use default config
	if len(args) == 0 {
		defaultComponents := o.getDefaultComponents()
		// update Components to default config
		o.Components = defaultComponents
		return nil
	}
	name := args[0]
	if v, ok := o.Components[name]; ok {
		if o.version == "" {
			o.Components = map[string]string{name: v}
		} else {
			o.Components = map[string]string{name: o.version}
		}
		return nil
	}
	return fmt.Errorf("component `%v` is not supported", name)
}

// Install component
func (o *InstallOptions) Install(component, version string) error {
	cmd, err := o.buildCmd(component, version)
	if err != nil {
		return err
	}
	return runCmd(cmd)
}

// buildCmd build cmd for installation
func (o *InstallOptions) buildCmd(component, version string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	var url string
	switch component {
	case "cert-manager":
		componentFile := "cert-manager.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", o.obUrl, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-operator", "ob-operator-dev":
		componentFile := "operator.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", o.obUrl, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "local-path-provisioner", "local-path-provisioner-dev":
		componentFile := "local-path-storage.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", o.localPathUrl, version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-dashboard":
		if err := addHelmRepo(); err != nil {
			return nil, err
		}
		if err := updateHelmRepo(); err != nil {
			return nil, err
		}
		versionFlag := fmt.Sprintf("--version=%s", version)
		cmd = exec.Command("helm", "install", "oceanbase-dashboard", "ob-operator/oceanbase-dashboard", versionFlag)
	default:
		return nil, fmt.Errorf("unknown component: %s", component)
	}
	return cmd, nil
}

// checkCertManager checks cert-manager in the environment
func checkCertManager() bool {
	cmd := exec.Command("kubectl", "get", "crds", "-o", "name", "|", "grep", "cert-manager")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}

	// Check if the output contains cert-manager resources
	expectedResources := []string{
		"challenges.acme.cert-manager.io",
		"orders.acme.cert-manager.io",
		"certificaterequests.cert-manager.io",
		"certificates.cert-manager.io",
		"clusterissuers.cert-manager.io",
		"issuers.cert-manager.io",
	}

	for _, resource := range expectedResources {
		if !bytes.Contains(out.Bytes(), []byte(resource)) {
			return false
		}
	}

	return true
}

func (o *InstallOptions) getDefaultComponents() map[string]string {
	defaultComponents := make(map[string]string) // Initialize the map
	var componentsList []string
	if !checkCertManager() {
		componentsList = []string{"cert-manager", "ob-operator", "ob-dashboard"}
	} else {
		componentsList = []string{"ob-operator", "ob-dashboard"}
	}
	for _, component := range componentsList {
		defaultComponents[component] = o.Components[component]
	}
	return defaultComponents
}

// addHelmRepo add ob-operator helm repo
func addHelmRepo() error {
	cmdAddRepo := exec.Command("helm", "repo", "add", "ob-operator", "https://oceanbase.github.io/ob-operator/")
	output, err := cmdAddRepo.CombinedOutput()
	if err != nil {
		return fmt.Errorf("adding repo failed: %s, %s", err, output)
	}
	return nil
}

// updateHelmRepo update ob-operator helm repo
func updateHelmRepo() error {
	cmdUpdateRepo := exec.Command("helm", "repo", "update", "ob-operator")
	output, err := cmdUpdateRepo.CombinedOutput()
	if err != nil {
		return fmt.Errorf("updating repo failed: %s, %s", err, output)
	}
	return nil
}

// runCmd run cmd for components' installation
func runCmd(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s command failed with error: %v", cmd.String(), err)
	}
	return nil
}
