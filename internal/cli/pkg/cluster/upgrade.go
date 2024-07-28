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

type UpgradeOptions struct {
	BaseOptions
	Image string `json:"image"`
}

func NewUpgradeOptions() *UpgradeOptions {
	return &UpgradeOptions{}
}
func (o *UpgradeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "namespace of ob cluster")
	cmd.Flags().StringVar(&o.Image, "image", "", "The image of observer")
}
func (o *UpgradeOptions) Validate() error {
	if o.Image == "" {
		return errors.New("image is required")
	}
	return nil
}
