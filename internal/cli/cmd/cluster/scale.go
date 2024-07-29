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

	apitypes "github.com/oceanbase/ob-operator/api/types"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/spf13/cobra"
)

// NewScaleCmd scale zones in ob cluster
func NewScaleCmd() *cobra.Command {
	o := cluster.NewScaleOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "scale <cluster_name>",
		Args:  cobra.ExactArgs(1),
		Short: "scale ob cluster",
		Run: func(cmd *cobra.Command, args []string) {
			o.Name = args[0]
			// TODO: support operation record
			if err := o.Parse(); err != nil {
				logger.Fatalln(err)
			}
			obcluster, err := clients.GetOBCluster(cmd.Context(), o.Namespace, o.Name)
			if err != nil {
				logger.Fatalln(err)
			}
			o.ZoneNum = len(obcluster.Spec.Topology)
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			if obcluster.Status.Status != clusterstatus.Running {
				logger.Fatalln(fmt.Errorf("Obcluster status invalid, Status:%s", obcluster.Status.Status))
			}
			for _, zone := range o.Topology {
				found := false
				for i := 0; i < len(obcluster.Spec.Topology); i++ {
					obzone := obcluster.Spec.Topology[i]
					if obzone.Zone == zone.Zone {
						found = true
						if zone.Replicas == 0 {
							obcluster.Spec.Topology = append(obcluster.Spec.Topology[:i], obcluster.Spec.Topology[i+1:]...)
							logger.Printf("Delete obzone %s", obzone.Zone)
						} else if obzone.Replica != zone.Replicas {
							obcluster.Spec.Topology[i].Replica = zone.Replicas
							logger.Printf("Scale obzone %s from %d to %d", obzone.Zone, obzone.Replica, zone.Replicas)
						} else {
							logger.Printf("No need to scale obzone %s", obzone.Zone)
						}
						break
					}
				}
				if !found {
					obcluster.Spec.Topology = append(obcluster.Spec.Topology, apitypes.OBZoneTopology{
						Zone:         zone.Zone,
						NodeSelector: common.KVsToMap(zone.NodeSelector),
						Replica:      zone.Replicas,
					})
				}
			}
			if err != nil {
				logger.Fatalln(err)
			}
			cluster, err := clients.UpdateOBCluster(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(oberr.NewInternal(err.Error()))
			}
			logger.Printf("Scale ob cluster %s success", cluster.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
