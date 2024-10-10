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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewListCmd list all ob tenants
func NewListCmd() *cobra.Command {
	o := tenant.NewListOptions()
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List ob tenants",
		Long:    `List ob tenants.`,
		Run: func(cmd *cobra.Command, args []string) {
			obtenantList, err := clients.ListAllOBTenants(cmd.Context(), o.Namespace, v1.ListOptions{})
			if err != nil {
				logger.Fatalln(err.Error())
			}
			sort.Slice(obtenantList.Items, func(i, j int) bool {
				return obtenantList.Items[i].Name < obtenantList.Items[j].Name
			})
			if len(obtenantList.Items) == 0 {
				logger.Println("No OBTenants found")
				return
			}
			tbLog.Println("NAMESPACE \t NAME \t CLUSTERNAME \t TENANTNAME \t TENANTROLE \t CREATETIME \t STATUS")
			for _, tenant := range obtenantList.Items {
				tbLog.Printf("%s \t %s \t %s \t %s \t %s \t %s \t %s\n", tenant.Namespace, tenant.ObjectMeta.Name, tenant.Spec.ClusterName, tenant.Spec.TenantName, tenant.Status.TenantRole, tenant.CreationTimestamp, tenant.Status.Status)
			}
			if err := tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	return cmd
}
