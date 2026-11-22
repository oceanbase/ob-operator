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

package v1alpha1

import (
	"errors"
	"fmt"
	"reflect"

	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
)

// ClassicBootstrapDispatchClaim identifies the live versioned task allowed to
// acquire the one-time SQL dispatch claim. The async worker can start before
// the coordinator persists its Running status.
func (r *OBCluster) ClassicBootstrapDispatchClaim() string {
	op := r.Status.OperationContext
	if r.UID == "" || r.Generation < 1 || r.DeletionTimestamp != nil || r.Spec.DeploymentMode == obconst.DeploymentModeSharedStorage ||
		r.Status.Status != clusterstatus.New || op == nil || op.Name != "bootstrap obcluster" ||
		op.Task != "dispatch classic bootstrap" || (op.TaskStatus != taskstatus.Pending && op.TaskStatus != taskstatus.Running) {
		return ""
	}
	return fmt.Sprintf("%s/%d", r.UID, r.Generation)
}

func (r *OBCluster) validateBootstrapClaimUpdate(old *OBCluster) error {
	before, existed := old.Annotations[obconst.AnnotationsClassicBootstrapClaim]
	after, exists := r.Annotations[obconst.AnnotationsClassicBootstrapClaim]
	if existed {
		if !exists || after != before {
			return errors.New("the bootstrap dispatch claim cannot be removed or changed")
		}
	} else if exists {
		expected := old.ClassicBootstrapDispatchClaim()
		if expected == "" || after != expected || old.UID != r.UID || old.Generation != r.Generation ||
			r.DeletionTimestamp != nil || !reflect.DeepEqual(old.Spec, r.Spec) {
			return errors.New("bootstrap dispatch claim does not match the current classic bootstrap task")
		}
	}
	return nil
}
