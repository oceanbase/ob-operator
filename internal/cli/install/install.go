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
	"os/exec"

	"github.com/spf13/cobra"

	utils "github.com/oceanbase/ob-operator/internal/cli/utils"
)

type InstallOptions struct {
	version    string
	name       string
	Components map[string]string
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{
		Components: utils.GetComponentsConf(),
	}
}

func (o *InstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.version, "version", "", "version of component")
	// cmd.Flags().StringToStringVar(&o.Components, "components", utils.GetComponentsConf(), "components config")
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
func (o *InstallOptions) Install() error {
	var url string
	baseUrl := "https://raw.githubusercontent.com/oceanbase/ob-operator/"
	for component, version := range o.Components {
		switch component {
		case "cert-manager":
			componentFile := "cert-manager.yaml"
			url = fmt.Sprintf("%s%s/deploy/%s", baseUrl, version, componentFile)
		case "ob-operator":
			componentFile := "operator.yaml"
			url = fmt.Sprintf("%s%s/deploy/%s", baseUrl, version, componentFile)
		}
		if err := run(url); err != nil {
			return err
		}
	}
	return nil
}

func run(url string) error {
	cmd := exec.Command("kubectl", "apply", "-f", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed with error: %v, output: %s", err, string(output))
	}
	return nil
}
