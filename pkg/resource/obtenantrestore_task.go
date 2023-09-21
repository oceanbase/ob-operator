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

package resource

import (
	"time"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

// Restore progress:
// 1. create unit (in tenant manager)
// 2. create resource pool (in tenant manager)
// 3. trigger restore job
// 4. wait for finishing
// 5. activate or replay log

// OBTenantManager tasks completion

func (m *OBTenantManager) CreateTenantRestoreJob() error {
	var existingJobs v1alpha1.OBTenantRestoreList
	var err error

	err = m.Client.List(m.Ctx, &existingJobs,
		client.MatchingLabels{
			oceanbaseconst.LabelRefOBCluster: m.OBTenant.Spec.ClusterName,
			oceanbaseconst.LabelTenantName:   m.OBTenant.Spec.TenantName,
		},
		client.InNamespace(m.OBTenant.Namespace))
	if err != nil {
		return err
	}

	if len(existingJobs.Items) != 0 {
		return errors.New("There is already at least one restore job for this tenant")
	}

	restoreJob := &v1alpha1.OBTenantRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.OBTenant.Spec.TenantName + "-restore",
			Namespace: m.OBTenant.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         m.OBTenant.APIVersion,
				Kind:               m.OBTenant.Kind,
				Name:               m.OBTenant.Name,
				UID:                m.OBTenant.GetUID(),
				BlockOwnerDeletion: getRef(true)}},
			Labels: map[string]string{
				oceanbaseconst.LabelRefOBCluster: m.OBTenant.Spec.ClusterName,
				oceanbaseconst.LabelTenantName:   m.OBTenant.Spec.TenantName,
				oceanbaseconst.LabelRefUID:       string(m.OBTenant.GetUID()),
			}},
		Spec: v1alpha1.OBTenantRestoreSpec{
			TargetTenant:  m.OBTenant.Spec.TenantName,
			TargetCluster: m.OBTenant.Spec.ClusterName,
			RestoreRole:   m.OBTenant.Spec.TenantRole,
			Source:        *m.OBTenant.Spec.Source,
		},
	}
	err = m.Client.Create(m.Ctx, restoreJob)
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) WatchRestoreJobToFinish() error {
	var err error
	for {
		runningRestore := &v1alpha1.OBTenantRestore{}
		err = m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBTenant.GetNamespace(),
			Name:      m.OBTenant.Spec.TenantName + "-restore",
		}, runningRestore)
		if err != nil {
			return err
		}
		if runningRestore.Status.Status == constants.RestoreJobSuccessful {
			break
		} else if runningRestore.Status.Status == constants.RestoreJobFailed {
			return errors.New("Restore job failed")
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

// OBTenantRestore tasks

func (m *ObTenantRestoreManager) StartRestoreJobInOB() error {
	return nil
}

func (m *ObTenantRestoreManager) CheckRestoreProgress() error {
	return nil
}

func (m *ObTenantRestoreManager) StartLogReplay() error {
	return nil
}

func (m *ObTenantRestoreManager) ActivateStandby() error {
	return nil
}

func (m *ObTenantRestoreManager) CancelRestoreJob() error {
	return nil
}

// get operation manager to exec sql
func (m *ObTenantRestoreManager) getOperationManager() (*operation.OceanbaseOperationManager, error) {
	if m.con != nil {
		return m.con, nil
	}
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.TargetCluster,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	m.con = con
	return con, nil
}
