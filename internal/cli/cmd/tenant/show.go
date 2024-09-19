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
package tenant

import (
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewShowCmd show the overview and operations of ob tenant
func NewShowCmd() *cobra.Command {
	o := tenant.NewShowOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	cmd := &cobra.Command{
		Use:     "show <tenant_name>",
		Short:   "Show overview of an ob tenant",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			obtenant, err := clients.GetOBTenant(cmd.Context(), types.NamespacedName{
				Namespace: o.Namespace,
				Name:      o.Name,
			})
			if err != nil {
				logger.Fatalln(err)
			}
			obtenantOperation, err := clients.GetOBTenantOperations(cmd.Context(), obtenant)
			if err != nil {
				logger.Fatalln(err)
			}
			tbLog.Println("TENANTNAME \t CLUSTERNAME \t TENANTROLE \t STATUS")
			tbLog.Printf("%s \t %s \t %s \t %s \n\n", obtenant.Spec.TenantName, obtenant.Spec.ClusterName, obtenant.Status.TenantRole, obtenant.Status.Status)
			if len(obtenant.Status.Pools) > 0 {
				tbLog.Println("ZONELIST \t UNITNUM \t PRIORITY")
				for _, pool := range obtenant.Status.Pools {
					tbLog.Printf("%s \t %d \t %d\n\n", pool.ZoneList, pool.UnitNumber, pool.Priority)
				}
			}
			if len(obtenantOperation.Items) > 0 {
				sort.Slice(obtenantOperation.Items, func(i, j int) bool {
					return obtenantOperation.Items[i].Name < obtenantOperation.Items[j].Name
				})
				tbLog.Println("OPERATION TYPE \t STATUS \t CREATETIME")
				for _, op := range obtenantOperation.Items {
					tbLog.Printf("%s \t %s \t %s\n", op.Spec.Type, op.Status.Status, op.CreationTimestamp)
				}
			} else {
				logger.Printf("No OBTenantOperations found in %s", obtenant.Spec.TenantName)
			}
			if err = tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
