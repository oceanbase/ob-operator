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
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewCreateCmd create an ob cluster
func NewCreateCmd() *cobra.Command {
	o := cluster.NewCreateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "create <cluster_name>",
		Short:   "Create ob cluster",
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.ClusterName = args[0]
			if err := o.Parse(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			obcluster := cluster.CreateOBClusterInstance(o)
			if err := clients.CreateSecretsForOBCluster(cmd.Context(), obcluster, o.RootPassword); err != nil {
				logger.Fatalln(errors.Wrap(err, "Create secrets for obcluster"))
			}
			_, err := clients.CreateOBCluster(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(oberr.NewInternal(err.Error()))
			}
			logger.Printf("Create obcluster instance: %s", o.ClusterName)
			logger.Printf("RootPassword: %s", o.RootPassword)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
