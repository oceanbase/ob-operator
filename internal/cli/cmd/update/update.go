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
package update

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/config"
	"github.com/oceanbase/ob-operator/internal/cli/update"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// NewCmd update the ob-operator and other components
func NewCmd() *cobra.Command {
	o := update.NewUpdateOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "update <component>",
		Short: "Command for ob-operator and other components update",
		Long: `Command for ob-operator and other components update.

Currently support:
- ob-operator: A Kubernetes operator that simplifies the deployment and management of OceanBase cluster and related resources on Kubernetes, support stable and develop version.
- ob-dashboard: A web application that provides resource management capabilities.
- local-path-provisioner: Provides a way for the Kubernetes users to utilize the local storage in each node, Storage of OceanBase cluster relies on it, which should be installed beforehand.
- cert-manager: Creates TLS certificates for workloads in Kubernetes and renews the certificates before they expire, ob-operator relies on it for certificate management, which should be installed beforehand.
		
if not specified, update ob-operator and ob-dashboard by default`,
		PreRunE:               o.Parse,
		ValidArgs:             config.ComponentUpdateList,
		DisableFlagsInUseLine: true,
		Args:                  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			componentCount := 0
			for component, version := range o.Components {
				if utils.CheckIfComponentExists(component) {
					componentCount++
					logger.Printf("Updating component %s, version %s\n", component, version)
					if err := o.Update(component, version); err != nil {
						logger.Fatalln(err)
					} else {
						logger.Printf("%s update successfully\n", component)
					}
				} else {
					logger.Printf("Component %s is not found\n", component)
				}
			}
			if componentCount == 0 {
				logger.Println("No components to update")
			}
		},
	}
	return cmd
}
