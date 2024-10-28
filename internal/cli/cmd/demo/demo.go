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
package demo

import "github.com/spf13/cobra"

// NewCmd create demo command for cluster creation
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo <subcommand>",
		Short: "deploy demo ob cluster in easier way",
		Long:  `deploy demo ob cluster in easier way, currently support single node and three node cluster`,
	}
	cmd.AddCommand(NewSingleNodeCmd())
	cmd.AddCommand(NewThreeNodeCmd())
	return cmd
}
