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
	"sort"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/cli/backup"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
)

// NewListCmd list all backup policies
func NewListCmd() *cobra.Command {
	o := backup.NewListOptions()
	tbw, tbLog := utils.GetTableLoggerInstance()
	logger := utils.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all backup policies",
		Long:  `List all backup policies`,
		Run: func(cmd *cobra.Command, args []string) {
			obBackupPolicyList, err := clients.ListAllTenantBackupPolicies(cmd.Context(), o.Namespace, v1.ListOptions{})
			if err != nil {
				logger.Fatalln(err)
			}
			sort.Slice(obBackupPolicyList.Items, func(i, j int) bool {
				return obBackupPolicyList.Items[i].Name < obBackupPolicyList.Items[j].Name
			})
			if len(obBackupPolicyList.Items) == 0 {
				logger.Println("No backup policies found")
				return
			}
			tbLog.Println("NAMESPACE \t POLICYNAME \t STATUS \t TENANTNAME \t NEXTFULL \t NEXTINCREMENTAL \t FULLCRONTAB \t INCREMENTALCRONTAB \t")
			for _, obBackupPolicy := range obBackupPolicyList.Items {
				tbLog.Printf("%s \t %s \t %s \t %s \t %s \t %s \t %s \t %s \n", obBackupPolicy.Namespace, obBackupPolicy.Name, obBackupPolicy.Status.Status, obBackupPolicy.Spec.TenantName, obBackupPolicy.Status.NextFull, obBackupPolicy.Status.NextIncremental, obBackupPolicy.Spec.DataBackup.FullCrontab, obBackupPolicy.Spec.DataBackup.IncrementalCrontab)
			}
			if err := tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
