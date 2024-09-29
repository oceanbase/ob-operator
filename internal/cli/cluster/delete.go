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
package cluster

import (
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/generic"
)

type DeleteOptions struct {
	generic.ResourceOption
}

func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// AddFlags add basic flags for cluster management
func (o *DeleteOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, "default", "namespace of ob cluster")
}
