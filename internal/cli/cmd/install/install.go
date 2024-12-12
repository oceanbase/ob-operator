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
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/install"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// NewCmd install the ob-operator and other components
func NewCmd() *cobra.Command {
	o := install.NewInstallOptions()
	logger := utils.GetDefaultLoggerInstance()
	componentList := []string{}
	for component := range o.Components {
		componentList = append(componentList, component)
	}
	cmd := &cobra.Command{
		Use:   "install <component>",
		Short: "Command for ob-operator and other components installation",
		Long: `Command for ob-operator and other components installation.

Currently support:
- ob-operator: A Kubernetes operator that simplifies the deployment and management of OceanBase cluster and related resources on Kubernetes, support stable and develop version.
- ob-dashboard: A web application that provides resource management capabilities.
- local-path-provisioner: Provides a way for the Kubernetes users to utilize the local storage in each node, Storage of OceanBase cluster relies on it, which should be installed beforehand, support stable and develop version.
- cert-manager: Creates TLS certificates for workloads in Kubernetes and renews the certificates before they expire, ob-operator relies on it for certificate management, which should be installed beforehand.

if not specified, install ob-operator and ob-dashboard by default, and cert-manager if it is not found in cluster.`,
		PreRunE:   o.Parse,
		ValidArgs: componentList,
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logger.Println("Install ob-operator and ob-dashboard by default")
			}
			for component, version := range o.Components {
				if err := o.Install(component, version); err != nil {
					logger.Fatalln(err)
				} else {
					logger.Printf("%s install successfully\n", component)
				}
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
