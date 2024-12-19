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
	"context"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

type ShowOptions struct {
	generic.ResourceOption
	jobType string
	limit   int64
}

func NewShowOptions() *ShowOptions {
	return &ShowOptions{}
}

func (o *ShowOptions) Parse(cmd *cobra.Command, args []string) error {
	o.Name = args[0]
	o.Cmd = cmd
	o.jobType = strings.ToUpper(o.jobType)
	return nil
}

func ListBackupJobs(ctx context.Context, policyName string, o *ShowOptions) (*v1alpha1.OBTenantBackupList, error) {
	listOption := metav1.ListOptions{}
	if o.jobType != "" && o.jobType != "ALL" {
		listOption.LabelSelector = oceanbaseconst.LabelRefBackupPolicy + "=" + policyName + "," + oceanbaseconst.LabelBackupType + "=" + o.jobType
	} else {
		listOption.LabelSelector = oceanbaseconst.LabelRefBackupPolicy + "=" + policyName
	}
	listOption.Limit = o.limit
	jobs, err := clients.ListBackupJobs(ctx, listOption)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// AddFlags adds flags for show command
func (o *ShowOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Namespace, FLAG_NAMESPACE, SHORTHAND_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of ob tenant, if not set, use default namespace")
	cmd.Flags().StringVarP(&o.jobType, FLAG_JOB_TYPE, SHORTHAND_TYPE, DEFAULT_JOBTYPE, "The type of backup job, support FULL, INC, CLEAN, ARCHIVE, ALL")
	cmd.Flags().Int64Var(&o.limit, FLAG_LIMIT, DEFAULT_LIMIT, "The number of backup jobs to show")
}
