/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obtenantrestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

const (
	restoreSubmission  = "oceanbase.oceanbase.com/restore-submission"
	trackedRestoreFlow = "start tracked restore"
)

func restoreDescription(r *v1alpha1.OBTenantRestore) string {
	return fmt.Sprintf("ob-operator:%s:%d", r.UID, r.Generation)
}

func (m *ObTenantRestoreManager) CanReplayTask(name tasktypes.TaskName) bool {
	oc := m.Resource.Status.OperationContext
	return name == tStartRestoreJobInOB && oc != nil && (oc.Name == trackedRestoreFlow || m.Resource.Annotations[restoreSubmission] == restoreDescription(m.Resource))
}

// Reserve this one physical submission before sending SQL. A replacement
// worker observes the UID-tagged job if the first worker's outcome was lost.
func (m *ObTenantRestoreManager) claimRestoreSubmission() (bool, error) {
	var claimed bool
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current v1alpha1.OBTenantRestore
		if err := m.Client.Get(m.Ctx, client.ObjectKeyFromObject(m.Resource), &current); err != nil {
			return err
		}
		if current.UID == "" || current.UID != m.Resource.UID || current.Generation != m.Resource.Generation || !current.DeletionTimestamp.IsZero() {
			return errors.New("restore resource was replaced, changed or deleted before submission")
		}
		description := restoreDescription(&current)
		if reserved, ok := current.Annotations[restoreSubmission]; ok {
			if reserved != description {
				return errors.New("restore submission belongs to another resource generation")
			}
			return nil
		}
		base := current.DeepCopy()
		if current.Annotations == nil {
			current.Annotations = map[string]string{}
		}
		current.Annotations[restoreSubmission] = description
		if err := m.Client.Patch(m.Ctx, &current, client.MergeFromWithOptions(base, client.MergeFromWithOptimisticLock{})); err != nil {
			return err
		}
		claimed = true
		return nil
	})
	return claimed, err
}

type trackedRestoreReader interface {
	GetTrackedRestoreJob(context.Context, string, string) (*model.RestoreHistory, error)
}

func (m *ObTenantRestoreManager) awaitSubmittedRestore(db trackedRestoreReader, description string) error {
	timeout := time.Duration(obcfg.GetConfig().Time.TenantRestoreTimeoutSeconds) * time.Second
	return waitForSubmittedRestore(m.Ctx, db, m.Resource.Spec.TargetTenant, description, timeout, 5*time.Second)
}
func waitForSubmittedRestore(ctx context.Context, db trackedRestoreReader, target, description string, timeout, interval time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := wait.PollUntilContextCancel(ctx, interval, true, func(ctx context.Context) (bool, error) {
		job, err := db.GetTrackedRestoreJob(ctx, target, description)
		if err != nil || job == nil {
			return false, err
		}
		if job.Description == nil || *job.Description != description || job.RestoreTenantName != target || job.JobID <= 0 {
			return false, errors.New("physical restore identity does not match")
		}
		if job.Status == "FAIL" || job.Status == "FAILED" {
			return false, fmt.Errorf("physical restore job %d failed", job.JobID)
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for physical restore of %s: %w", target, err)
	}
	return nil
}
