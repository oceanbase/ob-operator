/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package inspection

import (
	"context"
	"fmt"
	"reflect"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	insconst "github.com/oceanbase/ob-operator/internal/dashboard/business/inspection/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/config"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func inspectionClusterRoleName(cfg *config.Config) string {
	if cfg.Inspection.ClusterRoleName != "" {
		return cfg.Inspection.ClusterRoleName
	}
	// Compatibility with configurations from charts predating clusterRoleName.
	return insconst.ClusterRoleName
}

func ensureInspectionAccess(ctx context.Context, kube kubernetes.Interface, meta *response.OBClusterMetaBasic, roleName string) error {
	// Fail before creating a job or PVC if the configured role is unavailable.
	if _, err := kube.RbacV1().ClusterRoles().Get(ctx, roleName, metav1.GetOptions{}); err != nil {
		return errors.Wrapf(err, "Failed to get inspection cluster role %s", roleName)
	}
	saName := fmt.Sprintf(insconst.ServiceAccountNameFmt, meta.Name)
	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
		Name: saName, Namespace: meta.Namespace,
		OwnerReferences: []metav1.OwnerReference{{
			APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBCluster",
			Name: meta.Name, UID: types.UID(meta.UID),
		}},
	}}
	if _, err := kube.CoreV1().ServiceAccounts(meta.Namespace).Create(ctx, sa, metav1.CreateOptions{}); err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "Failed to create service account")
	}
	desired := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf(insconst.ClusterRoleBindingNameFmt, meta.Namespace, meta.Name)},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: saName, Namespace: meta.Namespace}},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: roleName},
	}
	bindings := kube.RbacV1().ClusterRoleBindings()
	existing, err := bindings.Get(ctx, desired.Name, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		_, err = bindings.Create(ctx, desired, metav1.CreateOptions{})
		if !k8serrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "Failed to create cluster role binding")
		}
		// Another request may have created it; verify rather than ignore it.
		existing, err = bindings.Get(ctx, desired.Name, metav1.GetOptions{})
	}
	if err != nil {
		return errors.Wrap(err, "Failed to get cluster role binding")
	}
	if !reflect.DeepEqual(existing.Subjects, desired.Subjects) {
		return errors.Errorf("Inspection cluster role binding %s has unexpected subjects", desired.Name)
	}
	if existing.RoleRef == desired.RoleRef {
		return nil
	}
	// RoleRef is immutable. Replace only the binding for this exact inspection
	// account, with preconditions protecting concurrent changes or replacement.
	if err := bindings.Delete(ctx, existing.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{
		UID: &existing.UID, ResourceVersion: &existing.ResourceVersion,
	}}); err != nil {
		return errors.Wrap(err, "Failed to replace inspection cluster role binding")
	}
	if _, err := bindings.Create(ctx, desired, metav1.CreateOptions{}); err != nil {
		return errors.Wrap(err, "Failed to create replacement inspection cluster role binding")
	}
	return nil
}
