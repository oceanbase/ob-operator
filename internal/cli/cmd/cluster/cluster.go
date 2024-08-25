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
package cluster

import (
	"github.com/spf13/cobra"
)

// NewCmd is command for cluster management
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster <subcommand>",
		Short: "Command for cluster management",
		Long:  `Command for cluster management, such as Create, UpGrade, Delete, Scale, Show.`,
	}
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewUpgradeCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewScaleCmd())
	cmd.AddCommand(NewShowCmd())
	return cmd
}
