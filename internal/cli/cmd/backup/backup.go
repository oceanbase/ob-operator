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
package backup

import "github.com/spf13/cobra"

// NewCmd is command for backup policy management
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup-policy <subcommand>",
		Short: "Command for backup policy management",
		Long:  `Command for backup policy management, such as create, list, delete, pause, resume, update`,
	}
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewPauseCmd())
	cmd.AddCommand(NewResumeCmd())
	cmd.AddCommand(NewShowCmd())
	return cmd
}
