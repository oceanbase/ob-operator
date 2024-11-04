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
package generic

import (
	"errors"

	"github.com/spf13/cobra"
)

type ResourceOption struct {
	Name      string
	Namespace string
	Cmd       *cobra.Command
}

// Parse the args in obocli
func (o *ResourceOption) Parse(cmd *cobra.Command, args []string) error {
	o.Name = args[0]
	o.Cmd = cmd
	return nil
}

// Complete the unset params in option
func (o *ResourceOption) Complete() error {
	return nil
}

// Validate the params in option
func (o *ResourceOption) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace is not specified")
	}
	return nil
}

// CheckIfFlagChanged checks if flags has changed
func (o *ResourceOption) CheckIfFlagChanged(flags ...string) bool {
	for _, flagName := range flags {
		if flag := o.Cmd.Flags().Lookup(flagName); flag != nil && flag.Changed {
			return true
		}
	}
	return false
}
