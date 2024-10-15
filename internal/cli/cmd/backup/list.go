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

	"github.com/oceanbase/ob-operator/internal/cli/backup"
	cmdUtil "github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewListCmd list all backup policies
func NewListCmd() *cobra.Command {
	o := backup.NewListOptions()
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	logger := cmdUtil.GetDefaultLoggerInstance()
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
			tbLog.Println("NAMESPACE \t POLICYNAME \t TENANTNAME \t JOBKEEP \t SUSPEND \t CREATETIME \t NEXTFULL \t NEXTINCREMENTAL \t STATUS")
			for _, policy := range obBackupPolicyList.Items {
				tbLog.Printf("%s \t %s \t %s \t %s \t %t \t %s \t %s \t %s \t %s\n", policy.Namespace, policy.Name, policy.Spec.TenantName, policy.Spec.JobKeepWindow, policy.Spec.Suspend, policy.CreationTimestamp, policy.Status.NextFull, policy.Status.NextIncremental, policy.Status.Status)
			}
			if err := tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
