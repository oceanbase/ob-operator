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

	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/config"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

type UpdateOptions struct {
	Components map[string]string
}

// NewUpdateOptions create a new UpdateOptions
func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		Components: make(map[string]string),
	}
}

func (o *UpdateOptions) Parse(_ *cobra.Command, args []string) error {
	// if specified, use the specified component
	if len(args) > 0 {
		name := args[0]
		components := config.GetAllComponents()
		// check if the component is supported
		defaultVersion, exist := components[name]
		if !exist {
			return fmt.Errorf("component %s is not supported", name)
		}
		o.Components = map[string]string{name: defaultVersion}
	} else {
		// if no component is specified, update default components
		defaultComponents := config.GetDefaultComponents()
		o.Components = defaultComponents
	}
	return nil
}

func (o *UpdateOptions) Update(component, version string) error {
	cmd, err := utils.BuildCmd(component, version)
	if err != nil {
		return err
	}
	return utils.RunCmd(cmd)
}
