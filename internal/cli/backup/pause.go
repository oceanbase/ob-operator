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

import "github.com/spf13/cobra"

type PauseOptions struct {
	UpdateOptions
}

func NewPauseOptions() *PauseOptions {
	return &PauseOptions{
		UpdateOptions: UpdateOptions{
			Status: "PAUSED",
		},
	}
}

func (o *PauseOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Name, FLAG_NAME, "", "The name of the tenant")
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of the tenant")
}
