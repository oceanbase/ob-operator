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

type ResourceOptions struct {
	Name      string
	Namespace string
}

// Parse the args in obocli
func (o *ResourceOptions) Parse(cmd *cobra.Command, args []string) error {
	o.Name = args[0]
	return nil
}

// Complete the unset params in options
func (o *ResourceOptions) Complete() error {
	return nil
}

// Validate the params in options
func (o *ResourceOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace not specified")
	}
	return nil
}

// AddFlags add basic flags for cluster management
func (o *ResourceOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of ob cluster")
}
