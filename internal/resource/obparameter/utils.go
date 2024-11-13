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

package obparameter

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func (m *OBParameterManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBParameter.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBParameterManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBParameter.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster")
	}
	return obcluster, nil
}

func (m *OBParameterManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
}

func (m *OBParameterManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		parameter := &v1alpha1.OBParameter{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBParameter.GetNamespace(),
			Name:      m.OBParameter.GetName(),
		}, parameter)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		parameter.Status = *m.OBParameter.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, parameter)
	})
}
