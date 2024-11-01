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

	"github.com/oceanbase/ob-operator/internal/cli/backup"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

// NewCreateCmd create an new backup policy
func NewCreateCmd() *cobra.Command {
	o := backup.NewCreateOptions()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "create <tenant_name>",
		Short:   "Create a backup policy for the specified ob tenant",
		Long:    `Create a backup policy for the specified ob tenant.`,
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(); err != nil {
				logger.Fatalln(err)
			}
			if err := o.Validate(); err != nil {
				logger.Fatalln(err)
			}
			obBackupPolicy, err := backup.CreateTenantBackupPolicy(cmd.Context(), o)
			if err != nil {
				logger.Fatalln(err)
			}
			logger.Printf("Create Backup policy %s for OBTenant %s successfully\n", obBackupPolicy.Name, o.Name)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
