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

	"k8s.io/apimachinery/pkg/types"
)

// NewShowCmd shows the backup policie and backup jobs of the tenant
func NewShowCmd() *cobra.Command {
	o := backup.NewShowOptions()
	tbw, tbLog := cmdUtil.GetTableLoggerInstance()
	logger := cmdUtil.GetDefaultLoggerInstance()
	cmd := &cobra.Command{
		Use:     "show <tenant-name>",
		Short:   "show backup policies and backup jobs of the ob tenant",
		Long:    `show backup policies and backup jobs of the ob tenant`,
		Args:    cobra.ExactArgs(1),
		PreRunE: o.Parse,
		Run: func(cmd *cobra.Command, args []string) {
			backupPolicy, err := clients.GetTenantBackupPolicy(cmd.Context(), types.NamespacedName{
				Name:      o.Name,
				Namespace: o.Namespace,
			})
			if err != nil {
				logger.Fatalln(err, "failed to get backup policy")
			}
			if backupPolicy == nil {
				logger.Fatalln("no backup policy found")
			}
			backupJobList, err := backup.ListBackupJobs(cmd.Context(), backupPolicy.Name, o)
			if err != nil {
				logger.Fatalln(err, "failed to list backup jobs")
			}
			tbLog.Println("POLICYNAME \t STATUS \t NEXTFULL \t NEXTINCREMENTAL \t FULLCRONTAB \t INCREMENTALCRONTAB \t")
			tbLog.Printf("%s \t %s \t %s \t %s \t %s \t %s \n\n", backupPolicy.Name, backupPolicy.Status.Status, backupPolicy.Status.NextFull, backupPolicy.Status.NextIncremental, backupPolicy.Spec.DataBackup.FullCrontab, backupPolicy.Spec.DataBackup.IncrementalCrontab)
			if len(backupJobList.Items) == 0 {
				logger.Println("no backup jobs found")
			}
			if len(backupJobList.Items) > 0 {
				sort.Slice(backupJobList.Items, func(i, j int) bool {
					return backupJobList.Items[i].CreationTimestamp.Before(&backupJobList.Items[j].CreationTimestamp)
				})
			}
			for _, job := range backupJobList.Items {
				tbLog.Println("JOBNAME \t STATUS \t TYPE \t STARTAT \t ENDAT \t")
				tbLog.Printf("%s \t %s \t %s \t %s \t %s \n", job.Name, job.Status.Status, job.Spec.Type, job.Status.StartedAt, job.Status.EndedAt)
			}
			if err = tbw.Flush(); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
