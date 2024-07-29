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
	"errors"

	"github.com/spf13/cobra"
)

type UpdateOptions struct {
	BaseOptions
	Cpu      int64 `json:"cpu"`
	MemoryGB int64 `json:"memoryGB"`
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{}
}
func (o *UpdateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of ob cluster")
	cmd.Flags().Int64Var(&o.Cpu, "cpu", 0, "The cpu of the observer")
	cmd.Flags().Int64Var(&o.MemoryGB, "memory", 0, "The memory of the observer")
}
func (o *UpdateOptions) Validate() error {
	if o.Cpu == 0 && o.MemoryGB == 0 {
		return errors.New("please specify update options, support cpu, memoryGB")
	}
	return nil
}
