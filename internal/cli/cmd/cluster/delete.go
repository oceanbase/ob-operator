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
	"errors"

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
)

func NewDeleteCmd() *cobra.Command {
	o := cluster.NewDeleteOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "delete <cluster_name>",
		Aliases: []string{"d"},
		Short:   "Delete ob cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logger.Println(errors.New("Cluster name is required"))
				return
			}
			o.Names = args
			for _, name := range o.Names {
				err := clients.DeleteOBCluster(cmd.Context(), o.Namespace, name)
				if err != nil {
					logger.Println(err)
				}
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
