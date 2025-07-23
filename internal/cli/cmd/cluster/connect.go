/*
Copyright (c) 2025 OceanBase
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

	cluster "github.com/oceanbase/ob-operator/internal/cli/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// NewConnectCmd connect to an ob cluster by sys tenant
func NewConnectCmd() *cobra.Command {
	o := cluster.NewConnectOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "connect <cluster_name>",
		Short:   "Connect to an ob cluster by sys tenant",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Run(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
