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
	if _, ok := o.Components[name]; !ok {
		return fmt.Errorf("%s install not supported", name)
	}
	if o.version != "" {
		o.Components[name] = o.version
	}
	return nil
}

// InstallAll install ob related components, and check the cert-manager in the environment
func (o *InstallOptions) InstallAll() error {
	if !checkCertManager() {
		if err := o.Install("cert-manager"); err != nil {
			return err
		}
	}
	if err := o.Install("ob-operator"); err != nil {
		return err
	}
	if err := o.Install("ob-dashboard"); err != nil {
		return err
	}
	return nil
}

// Install component
func (o *InstallOptions) Install(component string) error {
	var (
		cmd          *exec.Cmd
		url          string
		obUrl        string
		localPathUrl string
	)
	obUrl = "https://raw.githubusercontent.com/oceanbase/ob-operator/"
	localPathUrl = "https://raw.githubusercontent.com/rancher/local-path-provisioner/"
	switch component {
	case "cert-manager":
		componentFile := "cert-manager.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", obUrl, o.version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-operator", "ob-operator-dev":
		componentFile := "operator.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", obUrl, o.version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "local-path-provisioner", "local-path-provisioner-dev":
		componentFile := "local-path-storage.yaml"
		url = fmt.Sprintf("%s%s/deploy/%s", localPathUrl, o.version, componentFile)
		cmd = exec.Command("kubectl", "apply", "-f", url)
	case "ob-dashboard":
		if err := addHelmRepo(); err != nil {
			return err
		}
		if err := updateHelmRepo(); err != nil {
			return err
		}
		versionFlag := fmt.Sprintf("--version=%s", o.version)
		cmd = exec.Command("helm", "install", "oceanbase-dashboard", "ob-operator/oceanbase-dashboard", versionFlag)
	}
	return runCmd(cmd)
}

// checkCertManager checks cert-manager in the environment
func checkCertManager() bool {
	cmd := exec.Command("kubectl", "get", "crds", "-o", "name", "|", "grep", "cert-manager")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
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
		return fmt.Errorf("command failed with error: %v", err)
	}
	return nil
}
