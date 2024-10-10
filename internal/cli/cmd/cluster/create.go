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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cluster "github.com/oceanbase/ob-operator/internal/cli/cluster"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewCreateCmd create an ob cluster
func NewCreateCmd() *cobra.Command {
	o := cluster.NewCreateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "create <cluster_name>",
		Short:   "Create an ob cluster",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(); err != nil {
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
				logger.Fatalln(err)
			}
			logger.Printf("Create OBCluster instance: %s", o.ClusterName)
			logger.Printf("Run `echo $(kubectl get secret %s -o jsonpath='{.data.password}'|base64 --decode)` to get the secrets", obcluster.Spec.UserSecrets.Root)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
