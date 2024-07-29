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
	"github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
)

// NewShowCmd show the overview and operations of ob cluster
func NewShowCmd() *cobra.Command {
	o := cluster.NewShowOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	cmd := &cobra.Command{
		Use:   "show <cluster_name>",
		Short: "show overview of ob cluster",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.Name = args[0]
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			obclusterOperation, err := clients.ListOBClusterOperations(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(err)
			}
			o.Obcluster = obcluster
			o.ObclusterOperation = obclusterOperation
			// TODO: 显示operation信息
			tbLog.Println("Cluster ID \t Cluster Name \t Cluster Status \t Cluster Image")
			tbLog.Printf("%d \t %s \t %s \t %s \n", o.Obcluster.Spec.ClusterId, o.Obcluster.Spec.ClusterName, o.Obcluster.Status.Status, o.Obcluster.Status.Image)
			if len(o.Obcluster.Status.OBZoneStatus) > 0 {
				tbLog.Println("Zone \t Status")
				for _, zone := range o.Obcluster.Status.OBZoneStatus {
					tbLog.Printf("%s \t %s \n", zone.Zone, zone.Status)
				}
			}
			if len(o.Obcluster.Status.Parameters) > 0 {
				tbLog.Println("Parameters: Key \t Value")
				for _, Parameter := range o.Obcluster.Status.Parameters {
					tbLog.Printf("%s \t %s \n", Parameter.Name, Parameter.Value)
				}
			}
			if err = tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	return cmd
}
