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
package cli

import (
	"log"

	"github.com/oceanbase/ob-operator/internal/cli/cmd/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/install"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/version"
	"github.com/spf13/cobra"
)

// NewCliCmd return ob-operator cli
func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obocli",
		Short: "OceanBase Operator CLI",
		Long:  "OceanBase Operator CLI tool to manage OceanBase clusters, tenants, and backups.",
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().Changed("version") {
				versionCmd := version.NewVersionCmd()
				versionCmd.Run(cmd, args)
			} else {
				_ = cmd.Help()
			}
		},
	}
	cmd.AddCommand(version.NewVersionCmd())
	cmd.AddCommand(cluster.NewClusterCmd())
	cmd.AddCommand(tenant.NewTenantCmd())
	cmd.AddCommand(install.NewInstallCmd())
	cmd.Flags().BoolP("version", "v", false, "Print the version number of oceanbase cli")
	return cmd
}

func Execute() {
	rootCmd := NewCliCmd()
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
