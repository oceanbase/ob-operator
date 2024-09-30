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
	"errors"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

type ReplayLogOptions struct {
	generic.ResourceOption
	RestoreUntilOptions
	force bool
}
type RestoreUntilOptions struct {
	Timestamp string `json:"timestamp,omitempty" example:"2024-02-23 17:47:00"`
	Unlimited bool   `json:"unlimited,omitempty"`
}

func NewReplayLogOptions() *ReplayLogOptions {
	return &ReplayLogOptions{}
}

func GetReplayLogOperation(o *ReplayLogOptions) *v1alpha1.OBTenantOperation {
	replayLogOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-replay-log-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBTenantOp: o.Name},
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpReplayLog,
			ReplayUntil: &v1alpha1.RestoreUntilConfig{
				Timestamp: &o.Timestamp,
				Unlimited: o.Unlimited,
			},
			TargetTenant: &o.Name,
			Force:        o.force,
		},
	}
	return &replayLogOp
}

func (o *ReplayLogOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace is not specified")
	}
	if !o.Unlimited && o.Timestamp != "" {
		return errors.New("timestamp is required if the restore is limited")
	}
	return nil
}

// AddFlags add basic flags for tenant management
func (o *ReplayLogOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, "default", "The namespace of OBTenant")
	cmd.Flags().BoolVarP(&o.force, FLAG_FORCE, "f", false, "force operation")
	cmd.Flags().StringVar(&o.RestoreUntilOptions.Timestamp, FLAG_UNTIL_TIMESTAMP, "", "timestamp for obtenant restore,example: 2024-02-23 17:47:00")
	cmd.Flags().BoolVar(&o.RestoreUntilOptions.Unlimited, FLAG_UNLIMITED, true, "time limit for obtenant restore")
}
