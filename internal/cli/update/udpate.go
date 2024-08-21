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
package update

import (
	"fmt"

	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/spf13/cobra"
)

type UpdateOptions struct {
	Components map[string]string
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{}
}

func (o *UpdateOptions) Parse(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		o.Components = utils.GetComponentsConf()
		return nil
	}
	name := args[0]
	if v, ok := o.Components[name]; !ok {
		return fmt.Errorf("%s update not supported", name)
	} else {
		o.Components[name] = v
	}
	return nil
}

func (o *UpdateOptions) Update() error {
	return nil
}
