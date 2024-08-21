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
	"gopkg.in/yaml.v2"
)

type ComponentVersions struct {
	Components map[string]string `yaml:"components"`
}

// filePath for test
var filePath = "internal/cli/LATEST_VERSION.yaml"
var cv ComponentVersions

type InstallOptions struct {
	version    string
	Components map[string]string
}

func init() {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Errorf("Error reading LATEST_VERSION file: %v", err))
	}

	err = yaml.Unmarshal(data, &cv)
	// panic if file not exists
	if err != nil {
		panic(fmt.Errorf("Error decoding LATEST_VERSION file: %v", err))
	}
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{}
}

func (o *InstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.version, "version", "", "version of component")
}

func (o *InstallOptions) Parse(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		o.Components = cv.Components
		return nil
	}
	name := args[0]
	if v, ok := o.Components[name]; !ok {
		return fmt.Errorf("%s install not supported", name)
	} else {
		if o.version != "" {
			o.Components[name] = o.version
		} else {
			o.Components[name] = v
		}
		return nil
	}
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
	cmd := exec.Command("kubectl apply", "-f", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed with error: %v, output: %s", err, string(output))
	}
	return nil
}
