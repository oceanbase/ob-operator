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
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	cluster "github.com/oceanbase/ob-operator/internal/cli/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewUpdateCmd update obcluster
func NewUpdateCmd() *cobra.Command {
	o := cluster.NewUpdateOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "update <cluster_name>",
		Short:   "Update an ob cluster",
		Long:    "Update an ob cluster, support cpu/memory/storage",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					logger.Fatalf("OBCluster %s not found", o.Name)
				} else {
					logger.Fatalln(err)
				}
			}
			if err := utils.CheckClusterStatus(obcluster); err != nil {
				logger.Fatalln(err)
			}
			op := cluster.GetUpdateOperation(o)
			if _, err = clients.CreateOBClusterOperation(cmd.Context(), op); err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create update operation for OBCluster %s successfully", op.Spec.OBCluster)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
