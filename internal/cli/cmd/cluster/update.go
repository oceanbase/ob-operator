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

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	cluster "github.com/oceanbase/ob-operator/internal/cli/pkg/cluster"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/spf13/cobra"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

// NewUpdateCmd update obcluster
func NewUpdateCmd() *cobra.Command {
	o := cluster.NewUpdateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "update <cluster_name>",
		Short: "Update ob cluster",
		Long:  "Update ob cluster, support update cpu/memory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: support update storage
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
			if o.Cpu != 0 {
				obcluster.Spec.OBServerTemplate.Resource.Cpu = *apiresource.NewQuantity(o.Cpu, apiresource.DecimalSI)
			}
			if o.MemoryGB != 0 {
				obcluster.Spec.OBServerTemplate.Resource.Memory = *apiresource.NewQuantity(o.MemoryGB*constant.GB, apiresource.BinarySI)
			}
			cluster, err := clients.UpdateOBCluster(cmd.Context(), obcluster)
			if err != nil {
				logger.Fatalln(oberr.NewInternal(err.Error()))
			}
			updateOp := v1alpha1.OBClusterOperation{
				ObjectMeta: v1.ObjectMeta{
					Name:      o.Name + "-update-" + rand.String(6),
					Namespace: o.Namespace,
				},
				Spec: v1alpha1.OBClusterOperationSpec{
					OBCluster: o.Name,
					Type:      apiconst.ClusterOpTypeModifyOBServers,
					ModifyOBServers: &v1alpha1.ModifyOBServersConfig{Resource: &types.ResourceSpec{
						Cpu:    obcluster.Spec.MonitorTemplate.Resource.Cpu,
						Memory: obcluster.Spec.OBServerTemplate.Resource.Memory,
					}},
				},
			}
			_, err = clients.CreateOBClusterOperation(cmd.Context(), &updateOp)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Obcluster %s update success", cluster.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
