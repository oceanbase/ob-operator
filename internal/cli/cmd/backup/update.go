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
package backup

import (
	"github.com/oceanbase/ob-operator/internal/cli/backup"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

// NewUpdateCmd update backup policy of the ob tenant
func NewUpdateCmd() *cobra.Command {
	o := backup.NewUpdateOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "update <tenant_name>",
		Short:   "Update backup policy of the specified ob tenant",
		Long:    `Update backup policy of the specified ob tenant`,
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			nn := types.NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			}
			obtenant, err := clients.GetOBTenant(cmd.Context(), nn)
			if err != nil {
				logger.Fatalln(err)
			}
			if err := util.CheckTenantStatus(obtenant); err != nil {
				logger.Fatalln(err)
			}
			if err := backup.UpdateTenantBackupPolicy(cmd.Context(), o); err != nil {
				logger.Fatalln(err)
			}
			logger.Println("update backup policy of the specified ob tenant")
		},
	}
	o.AddFlags(cmd)
	return cmd
}
