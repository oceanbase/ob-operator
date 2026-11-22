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
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type restoreRead func(context.Context, string, string) (*model.RestoreHistory, error)

func (f restoreRead) GetTrackedRestoreJob(c context.Context, n, d string) (*model.RestoreHistory, error) {
	return f(c, n, d)
}

func TestReservePhysicalRestoreOnce(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	r := &v1alpha1.OBTenantRestore{ObjectMeta: metav1.ObjectMeta{Namespace: "n", Name: "r", UID: "first-uid", Generation: 1}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).Build()
	manager := &ObTenantRestoreManager{Ctx: context.Background(), Client: cl, Resource: r.DeepCopy()}
	for _, want := range []bool{true, false, false} {
		first, err := manager.claimRestoreSubmission()
		if err != nil || first != want {
			t.Fatalf("acquired=%v want %v: %v", first, want, err)
		}
	}
	manager.Resource.UID = "other-uid"
	if got, err := manager.claimRestoreSubmission(); err == nil || got {
		t.Fatal("claimed replaced CR")
	}
	manager.Resource.UID = r.UID
	manager.Resource.Generation++
	if got, err := manager.claimRestoreSubmission(); err == nil || got {
		t.Fatal("claimed new spec generation")
	}
}

func TestReadOnlyRestoreAdoption(t *testing.T) {
	const tag = "ob-operator:the-current-uid:1"
	for _, tc := range []struct {
		name, description, target, status string
		exists, succeeds                  bool
	}{
		{"active", tag, "tenant", "RESTORING", true, true},
		{"already finished", tag, "tenant", "SUCCESS", true, true},
		{"old UID", "ob-operator:other:1", "tenant", "SUCCESS", true, false},
		{"unmarked history", "", "tenant", "SUCCESS", true, false},
		{"other target", tag, "another", "SUCCESS", true, false},
		{"physical failure", tag, "tenant", "FAIL", true, false},
		{"no evidence", tag, "tenant", "", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := restoreRead(func(_ context.Context, n, d string) (*model.RestoreHistory, error) {
				if n != "tenant" || d != tag {
					t.Fatal("incorrect job identity")
				}
				if !tc.exists {
					return nil, nil
				}
				return &model.RestoreHistory{RestoreProgress: model.RestoreProgress{RestoreTenantName: tc.target, JobID: 3, Description: &tc.description, Status: tc.status}}, nil
			})
			err := waitForSubmittedRestore(context.Background(), probe, "tenant", tag, 50*time.Millisecond, time.Millisecond)
			if (err == nil) != tc.succeeds {
				t.Fatalf("unexpected result: %v", err)
			}
			if !tc.exists && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("missing timeout: %v", err)
			}
		})
	}
}

func TestRestoreReplayIsScoped(t *testing.T) {
	r := &v1alpha1.OBTenantRestore{ObjectMeta: metav1.ObjectMeta{UID: "uid", Generation: 1}}
	m := &ObTenantRestoreManager{Resource: r}
	r.Status.OperationContext = &tasktypes.OperationContext{Name: "start restore"}
	if m.CanReplayTask(tStartRestoreJobInOB) {
		t.Fatal("old unmarked submission cannot be replayed safely")
	}
	r.Status.OperationContext.Name = trackedRestoreFlow
	if !m.CanReplayTask(tStartRestoreJobInOB) || m.CanReplayTask(tActivateStandby) {
		t.Fatal("replay policy should cover only tracked submission")
	}
}
