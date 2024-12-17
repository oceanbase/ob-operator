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
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/cmd/backup"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/completion"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/demo"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/install"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/tenant"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/update"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/version"
)

// BinaryName injected by ldflags
var BinaryName = "unknown"

var rootLongDesc = `         
=============================================
          _             _     _ 
   ___   | | __   ___  | |_  | |
  / _ \  | |/ /  / __| | __| | |
 | (_) | |   <  | (__  | |_  | |
  \___/  |_|\_\  \___|  \__| |_|

=============================================
A Command Line Tool compatible with OceanBase Operator`

// NewCliCmd return ob-operator cli
func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   BinaryName,
		Short: "OceanBase Operator Cli",
		Long:  rootLongDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("version") {
				versionCmd := version.NewCmd()
				return versionCmd.RunE(cmd, args)
			}
			return cmd.Help()
		},
	}
	cmd.AddCommand(version.NewCmd())
	cmd.AddCommand(cluster.NewCmd())
	cmd.AddCommand(tenant.NewCmd())
	cmd.AddCommand(backup.NewCmd())
	cmd.AddCommand(install.NewCmd())
	cmd.AddCommand(update.NewCmd())
	cmd.AddCommand(demo.NewCmd())
	cmd.AddCommand(completion.NewCmd(cmd.OutOrStdout(), ""))
	cmd.Flags().BoolP("version", "v", false, "Print the version of oceanbase cli")
	return cmd
}
