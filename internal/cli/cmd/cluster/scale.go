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
	"fmt"

	"github.com/spf13/cobra"

	cluster "github.com/oceanbase/ob-operator/internal/cli/cluster"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
)

// NewScaleCmd scale zones in ob cluster
func NewScaleCmd() *cobra.Command {
	o := cluster.NewScaleOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "scale <cluster_name>",
		Args:    cobra.ExactArgs(1),
		Short:   "Scale ob cluster, support add/adjust/delete of zones",
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			if obcluster.Status.Status != clusterstatus.Running {
				logger.Fatalln(fmt.Errorf("Obcluster status invalid, Status:%s", obcluster.Status.Status))
			}
			o.OldTopology = obcluster.Spec.Topology
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			op := cluster.GetScaleOperation(o)
			_, err = clients.CreateOBClusterOperation(cmd.Context(), op)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create scale operation for obcluster %s success", op.Spec.OBCluster)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
