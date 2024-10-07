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

	cluster "github.com/oceanbase/ob-operator/internal/cli/cluster"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewUpgradeCmd upgrade obclusters
func NewUpgradeCmd() *cobra.Command {
	o := cluster.NewUpgradeOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "upgrade <cluster_name>",
		Short:   "Upgrade an OBCluster",
		Long:    "Upgrade an OBCluster, please specify the new image",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"ug"},
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			if err := cmdUtil.CheckClusterStatus(obcluster); err != nil {
				logger.Fatalln(err)
			}
			op := cluster.GetUpgradeOperation(o)
			if _, err = clients.CreateOBClusterOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create upgrade operation for OBCluster %s successfully", op.Spec.OBCluster)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
