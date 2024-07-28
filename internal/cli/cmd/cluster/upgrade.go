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

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/spf13/cobra"
)

// NewUpgradeCmd upgrade obclusters
func NewUpgradeCmd() *cobra.Command {
	o := cluster.NewUpgradeOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "upgrade <cluster_name>",
		Short: "Upgrade ob cluster",
		Long:  "Upgrade ob cluster, please specify the new image",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.Name = args[0]
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			if obcluster.Status.Status != clusterstatus.Running {
				logger.Fatalln(fmt.Errorf("Obcluster status invalid, Status:%s", obcluster.Status.Status))
			}
			obcluster.Spec.OBServerTemplate.Image = o.Image
			cluster, err := clients.UpdateOBCluster(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(oberr.NewInternal(err.Error()))
			}
			logger.Printf("Obcluster %s update success", cluster.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
