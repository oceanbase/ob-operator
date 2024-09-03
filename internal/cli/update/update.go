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

	"github.com/oceanbase/ob-operator/internal/cli/install"
)

type UpdateOptions struct {
	install.InstallOptions
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		InstallOptions: *install.NewInstallOptions(),
	}
}

func (o *UpdateOptions) Parse(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	name := args[0]
	if _, ok := o.Components[name]; !ok {
		return fmt.Errorf("%s update not supported", name)
	}
	return nil
}
