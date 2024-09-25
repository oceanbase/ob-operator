/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obtenantoperation

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// get operation manager to exec sql
func (m *ObTenantOperationManager) getTenantRootClient(tenantName string) (*operation.OceanbaseOperationManager, error) {
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      tenantName,
	}, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get tenant")
	}
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      tenant.Spec.ClusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	var con *operation.OceanbaseOperationManager
	con, err = resourceutils.GetTenantRootOperationClient(m.Client, m.Logger, obcluster, tenant.Spec.TenantName, tenant.Status.Credentials.Root)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	return con, nil
}

func (m *ObTenantOperationManager) getClusterSysClient(clusterName string) (*operation.OceanbaseOperationManager, error) {
	var err error
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      clusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster sys client")
	}
	return con, nil
}

func (m *ObTenantOperationManager) retryUpdateTenant(obj *v1alpha1.OBTenant) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		tenant := &v1alpha1.OBTenant{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      obj.Name,
		}, tenant)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		tenant.Status = obj.Status
		return m.Client.Status().Update(m.Ctx, tenant)
	})
}

type tenantConditionMatcher func(t *v1alpha1.OBTenant) bool

func (m *ObTenantOperationManager) waitForOBTenantToBeStatus(waitSeconds int, matcher tenantConditionMatcher) error {
	if m.Resource.Spec.TargetTenant == nil {
		return errors.New("target tenant is nil")
	}
	for i := 0; i < waitSeconds; i++ {
		time.Sleep(time.Second)
		tenant := &v1alpha1.OBTenant{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      *m.Resource.Spec.TargetTenant,
		}, tenant)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		if matcher(tenant) {
			return nil
		}
	}
	return errors.Errorf("wait for tenant %s to be in desired status timeout", m.Resource.Name)
}
