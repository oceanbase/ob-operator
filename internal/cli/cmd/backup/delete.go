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
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/cli/backup"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewDeleteCmd delete backup policy
func NewDeleteCmd() *cobra.Command {
	o := backup.NewDeleteOptions()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "delete <tenant_name>",
		Short:   "Delete backup policy for the specified tenant",
		Long:    "Delete backup policy for the specified tenant",
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			err := clients.DeleteTenantBackupPolicy(cmd.Context(), types.NamespacedName{
				Namespace: o.Namespace,
				Name:      o.Name + "-backup-policy",
			})
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Delete backup policy for OBTenant %s successfully", o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
