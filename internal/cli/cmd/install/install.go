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

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/install"
)

// NewCmd install the ob-operator and other components
func NewCmd() *cobra.Command {
	o := install.NewInstallOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "install <components>",
		Short:   "Command for ob-operator and components installation",
		Long:    "Command for ob-operator and components installation, support ob-operator, ob-dashboard, local-path-provisioner, cert-manager",
		PreRunE: o.Parse,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logger.Println("Install all the components")
			}
			for component, version := range o.Components {
				if err := install.Install(component, version); err != nil {
					logger.Fatalln(err)
				} else {
					logger.Printf("%s install success", component)
				}
			}

		},
	}
	return cmd
}
