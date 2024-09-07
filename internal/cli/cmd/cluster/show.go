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
	"sort"

	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewShowCmd show the overview and operations of ob cluster
func NewShowCmd() *cobra.Command {
	o := cluster.NewShowOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	cmd := &cobra.Command{
		Use:     "show <cluster_name>",
		Short:   "Show overview of ob cluster",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			obclusterOperation, err := clients.GetOBClusterOperations(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(err)
			}
			tbLog.Println("ClUSTER ID \t NAME \t STATUS \t IMAGE")
			tbLog.Printf("%d \t %s \t %s \t %s \n\n", obcluster.Spec.ClusterId, obcluster.Spec.ClusterName, obcluster.Status.Status, obcluster.Status.Image)
			if len(obcluster.Status.OBZoneStatus) > 0 {
				tbLog.Println("ZONE \t STATUS")
				for _, zone := range obcluster.Status.OBZoneStatus {
					tbLog.Printf("%s \t %s \n\n", zone.Zone, zone.Status)
				}
			}
			if len(obcluster.Status.Parameters) > 0 {
				tbLog.Println("KEY \t VALUE")
				for _, Parameter := range obcluster.Status.Parameters {
					tbLog.Printf("%s \t %s \n\n", Parameter.Name, Parameter.Value)
				}
			}

			if len(obclusterOperation.Items) > 0 {
				sort.Slice(obclusterOperation.Items, func(i, j int) bool {
					return obclusterOperation.Items[i].Name < obclusterOperation.Items[j].Name
				})
				tbLog.Println("OPERATION TYPE \t TTLDAYS \t STATUS \t CREATETIME")
				for _, op := range obclusterOperation.Items {
					tbLog.Printf("%s \t %d \t  %s \t %s\n", op.Spec.Type, op.Spec.TTLDays, op.Status.Status, op.CreationTimestamp)
				}
			} else {
				logger.Printf("No OBClusterOperations found in %s", obcluster.Spec.ClusterName)
			}
			if err = tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
