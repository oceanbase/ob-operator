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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

type UpgradeOptions struct {
	generic.ResourceOption
	Image string `json:"image"`
}

func NewUpgradeOptions() *UpgradeOptions {
	return &UpgradeOptions{}
}

// GetUpgradeOperation creates upgrade opertaions
func GetUpgradeOperation(o *UpgradeOptions) *v1alpha1.OBClusterOperation {
	upgradeOp := &v1alpha1.OBClusterOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-upgrade-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBClusterOp: o.Name},
		},
		Spec: v1alpha1.OBClusterOperationSpec{
			OBCluster: o.Name,
			Type:      apiconst.ClusterOpTypeUpgrade,
			Upgrade:   &v1alpha1.UpgradeConfig{Image: o.Image},
		},
	}
	return upgradeOp
}

func (o *UpgradeOptions) Validate() error {
	if o.Image == "" {
		return errors.New("image is not specified")
	}
	return nil
}

// AddFlags for upgrade options
func (o *UpgradeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Namespace, FLAG_NAMESPACE, SHORTHAND_NAMESPACE, DEFAULT_NAMESPACE, "namespace of ob cluster")
	cmd.Flags().StringVar(&o.Image, FLAG_OBSERVER_IMAGE, "", "The image of observer") // set image to null, avoid image downgrade
	_ = cmd.MarkFlagRequired(FLAG_OBSERVER_IMAGE)
}
