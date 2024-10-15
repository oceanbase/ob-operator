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
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	generic.ResourceOption
}

func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// AddFlags adds flags for list command
func (o *ListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of ob tenant, if not set, use default namespace")
}
