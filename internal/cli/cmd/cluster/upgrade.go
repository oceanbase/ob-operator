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
	"fmt"
	"log"

	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/spf13/cobra"
)

func NewUpgradeCmd() *cobra.Command {
	o := cluster.NewUpgradeOptions()
	cmd := &cobra.Command{
		Use:   "upgrade <cluster_name>",
		Short: "Upgrade ob cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Println(errors.New("please specify cluster name"))
			}
			o.Names = args
			for _, name := range o.Names {
				obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, name)
				if err != nil {
					log.Println(err)
				}
				if obcluster.Status.Status != clusterstatus.Running {
					log.Println(fmt.Errorf("Obcluster status invalid %s", obcluster.Status.Status))
				}
				obcluster.Spec.OBServerTemplate.Image = o.Image
				cluster, err := clients.UpdateOBCluster(cmd.Context(), obcluster)
				if err != nil {
					log.Println(oberr.NewInternal(err.Error()))
				}
				log.Println(cluster)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
