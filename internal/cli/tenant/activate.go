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
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

type ActivateOptions struct {
	generic.ResourceOption
	force bool
}

func NewActivateOptions() *ActivateOptions {
	return &ActivateOptions{}
}

func GetActivateOperation(o *ActivateOptions) *v1alpha1.OBTenantOperation {
	activateOp := &v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-activate-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBTenantOp: o.Name},
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpFailover,
			Failover: &v1alpha1.OBTenantOpFailoverSpec{
				StandbyTenant: o.Name,
			},
			Force: o.force,
		},
	}
	return activateOp
}

// AddFlags add basic flags for tenant management
func (o *ActivateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Namespace, FLAG_NAMESPACE, SHORTHAND_NAMESPACE, DEFAULT_NAMESPACE, "namespace of ob tenant")
	cmd.Flags().BoolVarP(&o.force, FLAG_FORCE, SHORTHAND_FORCE, DEFAULT_FORCE_FLAG, "If the operation is a force operation")
}
