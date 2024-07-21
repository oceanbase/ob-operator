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
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var tbw = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
var tbLog = log.New(tbw, "", 0)

func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster <subcommand>",
		Short: "Command for cluster management",
		Long:  `Command for cluster management, such as Create, Update, Delete.`,
	}
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewUpgradeCmd())
	cmd.AddCommand(NewListCmd())
	return cmd
}
