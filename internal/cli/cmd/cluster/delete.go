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
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
)

// NewDeleteCmd delete ob clusters
func NewDeleteCmd() *cobra.Command {
	o := cluster.NewDeleteOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "delete <cluster_name>",
		Short:   "Delete ob cluster",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.Name = args[0]
			err := clients.DeleteOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Delete ob cluster %s success", o.Name)

		},
	}
	o.AddFlags(cmd)
	return cmd
}
