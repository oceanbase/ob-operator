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
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/clients"
)

type DeleteOptions struct {
	generic.ResourceOption
	force bool
}

func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

func DeleteTenantBackupPolicy(ctx context.Context, o *DeleteOptions) error {
	nn := types.NamespacedName{Name: o.Name, Namespace: o.Namespace}
	policy, err := clients.GetTenantBackupPolicy(ctx, nn)
	if policy == nil {
		return fmt.Errorf("Backup policy for %s not found", o.Name)
	}
	if err != nil {
		return err
	}
	if o.force {
		return clients.ForceDeleteTenantBackupPolicy(ctx, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace})
	}
	return clients.DeleteTenantBackupPolicy(ctx, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace})
}

// AddFlags add basic flags for tenant management
func (o *DeleteOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Namespace, FLAG_NAMESPACE, SHORTHAND_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of the ob tenant")
	cmd.Flags().BoolVarP(&o.force, FLAG_FORCE, SHORTHAND_FORCE, DEFAULT_FORCE, "Force delete the ob tenant backup policy")
}
