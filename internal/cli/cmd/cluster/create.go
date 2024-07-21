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
	"log"
	"time"

	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	o := cluster.NewCreateOptions()
	cmd := &cobra.Command{
		Use:   "create <cluster_name>",
		Short: "Create ob cluster",
		Run: func(cmd *cobra.Command, args []string) {
			// 获取当前时间并取模
			o.ClusterId = time.Now().Unix() % 4294901759
			o.ClusterName = args[0]
			obcluster := cluster.CreateOBClusterInstance(o)
			if obcluster == nil {
				return
			}
			err := clients.CreateSecretsForOBCluster(cmd.Context(), obcluster, o.RootPassword)
			if err != nil {
				log.Println(errors.Wrap(err, "Create secrets for ob cluster"))
			}
			cluster, err := clients.CreateOBCluster(cmd.Context(), obcluster)
			if err != nil {
				log.Println(errors.Wrap(err, "Create ob cluster").Error())
			}
			log.Println("Create obcluster instance: %v", cluster)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
