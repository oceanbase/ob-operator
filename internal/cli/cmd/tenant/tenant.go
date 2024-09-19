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
package tenant

import (
	"github.com/spf13/cobra"
)

// NewCmd is command for tenant management
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tenant",
		Aliases: []string{"t"},
		Short:   "Command for tenant management",
		Long:    `Command for tenant management, such as Create, Update, Delete, Switchover, Activate, Replaylog.`,
	}
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewScaleCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewUpgradeCmd())
	cmd.AddCommand(NewChangePwdCmd())
	cmd.AddCommand(NewSwitchOverCmd())
	cmd.AddCommand(NewActivateCmd())
	cmd.AddCommand(NewShowCmd())
	cmd.AddCommand(NewReplayLogCmd())
	return cmd
}
