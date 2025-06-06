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
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewShowCmd show the overview and operations of ob cluster
func NewShowCmd() *cobra.Command {
	o := cluster.NewShowOptions()
	logger := utils.GetDefaultLoggerInstance()
	tbw, tbLog := utils.GetTableLoggerInstance()
	cmd := &cobra.Command{
		Use:     "show <cluster_name>",
		Short:   "Show overview of an ob cluster",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					logger.Fatalf("OBCluster %s not found", o.Name)
				} else {
					logger.Fatalln(err)
				}
			}
			obclusterOperationList, err := clients.GetOBClusterOperations(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(err)
			}
			tbLog.Println("ClUSTER ID \t NAME \t STATUS \t IMAGE")
			tbLog.Printf("%d \t %s \t %s \t %s \n\n", obcluster.Spec.ClusterId, obcluster.Spec.ClusterName, obcluster.Status.Status, obcluster.Status.Image)
			if len(obcluster.Status.OBZoneStatus) > 0 {
				tbLog.Println("ZONE \t STATUS")
				for _, zone := range obcluster.Status.OBZoneStatus {
					tbLog.Printf("%s \t %s \n", zone.Zone, zone.Status)
				}
				tbLog.Println()
			}
			if len(obcluster.Status.Parameters) > 0 {
				tbLog.Println("KEY \t VALUE")
				for _, Parameter := range obcluster.Status.Parameters {
					tbLog.Printf("%s \t %s \n", Parameter.Name, Parameter.Value)
				}
				tbLog.Println()
			}

			if len(obclusterOperationList.Items) > 0 {
				sort.Slice(obclusterOperationList.Items, func(i, j int) bool {
					return obclusterOperationList.Items[i].Name < obclusterOperationList.Items[j].Name
				})
				tbLog.Println("OPERATION TYPE \t TTLDAYS \t STATUS \t CREATETIME")
				for _, op := range obclusterOperationList.Items {
					tbLog.Printf("%s \t %d \t  %s \t %s \n", op.Spec.Type, op.Spec.TTLDays, op.Status.Status, op.CreationTimestamp)
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
