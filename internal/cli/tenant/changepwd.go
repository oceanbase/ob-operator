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
	"context"
	"errors"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type ChangePwdOptions struct {
	generic.ResourceOption
	Password       string `json:"password" binding:"required"`
	RootSecretName string `json:"rootSecretName" binding:"required"`
	force          bool
}

func NewChangePwdOptions() *ChangePwdOptions {
	return &ChangePwdOptions{}
}

func GetChangePwdOperation(o *ChangePwdOptions) *v1alpha1.OBTenantOperation {
	changePwdOp := &v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.Name + "-change-root-pwd-" + rand.String(6),
			Namespace: o.Namespace,
			Labels:    map[string]string{oceanbaseconst.LabelRefOBTenantOp: o.Name},
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpChangePwd,
			ChangePwd: &v1alpha1.OBTenantOpChangePwdSpec{
				Tenant:    o.Name,
				SecretRef: o.RootSecretName,
			},
			Force: o.force,
		},
	}
	return changePwdOp
}

// GenerateNewPwd generate new password for obtenant
func GenerateNewPwd(ctx context.Context, o *ChangePwdOptions) error {
	k8sclient := client.GetClient()
	o.RootSecretName = o.Name + "-root-" + rand.String(6)
	_, err := k8sclient.ClientSet.CoreV1().Secrets(o.Namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      o.RootSecretName,
			Namespace: o.Namespace,
		},
		StringData: map[string]string{
			"password": o.Password,
		},
	}, v1.CreateOptions{})
	if err != nil {
		return oberr.NewInternal(err.Error())
	}
	return nil
}

func (o *ChangePwdOptions) Validate() error {
	if o.Password == "" {
		return errors.New("Password can not be empty")
	}
	return nil
}

func (o *ChangePwdOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "namespace of ob tenant")
	cmd.Flags().StringVarP(&o.Password, FLAG_PASSWD, "p", "", "new password of ob tenant")
	cmd.Flags().BoolVarP(&o.force, FLAG_FORCE, "f", DEFAULT_FORCE_FLAG, "force operation")
}
