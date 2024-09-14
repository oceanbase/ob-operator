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
package tenant

import (
	"errors"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

type UpdateOptions struct {
	generic.ResourceOptions
	force            bool
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`
	Charset          string `json:"charset,omitempty"`
	UnitNumber       int    `json:"unitNum" binding:"required"`
	UpdateType       string `json:"updateType"`
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{}
}

func GetUpdateOperations(o *UpdateOptions) *v1alpha1.OBTenantOperation {
	updateOp := &v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-update-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBTenantOp: o.Name},
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type:         apiconst.TenantOpSetCharset,
			TargetTenant: &o.Name,
			Force:        o.force,
		},
	}
	switch o.UpdateType {
	case "charset":
		updateOp.Spec.Charset = o.Charset
	case "connect-while-list":
		updateOp.Spec.ConnectWhiteList = o.ConnectWhiteList
	case "unit-number":
		updateOp.Spec.UnitNumber = o.UnitNumber
	}
	return updateOp
}

func (o *UpdateOptions) Validate() error {
	updateTypeCount := 0
	if o.Charset != "" {
		updateTypeCount++
		o.UpdateType = "charset"
	}
	if o.ConnectWhiteList != "" {
		updateTypeCount++
		o.UpdateType = "connect-white-list"
	}
	if o.UnitNumber != 0 {
		updateTypeCount++
		o.UpdateType = "unit-number"
	}
	if updateTypeCount > 1 {
		return errors.New("Only one type of update is allowed at a time")
	}
	if updateTypeCount == 0 {
		return errors.New("No update type specified, support cpu/memory/storage")
	}
	return nil
}

// AddFlags add basic flags for tenant management
func (o *UpdateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, "namespace", "default", "The namespace of OBTenant")
	cmd.Flags().IntVar(&o.UnitNumber, "unit-number", 0, "unit number of the OBTenant")
	cmd.Flags().StringVar(&o.Charset, "charset", "", "The charset using in ob tenant")
	cmd.Flags().StringVar(&o.ConnectWhiteList, "connect-white-list", "", "The connect white list using in ob tenant")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "force operation")
}
