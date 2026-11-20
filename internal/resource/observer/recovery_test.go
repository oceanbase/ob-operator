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

package observer

import (
	"context"
	"errors"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"

	"github.com/oceanbase/ob-operator/pkg/coordinator"
	tt "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestObserverReplayPolicy(t *testing.T) {
	var policy coordinator.TaskReplayPolicy = &OBServerManager{}
	for _, name := range []tt.TaskName{tWaitOBServerReady, tCreateOBServerPVC, tWaitOBClusterBootstrapped} {
		if !policy.CanReplayTask(name) {
			t.Fatalf("bootstrap task %q must be recoverable", name)
		}
	}
	for _, name := range []tt.TaskName{"", "bootstrap", "create obzone", "add server", tCreateOBServerSvc, tCreateOBServerPod, "unknown"} {
		if policy.CanReplayTask(name) {
			t.Fatalf("unsafe task %q allowed", name)
		}
	}
}

func TestLostParentBootstrapWaitNeverMarksObserverRunning(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{clusterstatus.New, clusterstatus.Bootstrapped, clusterstatus.Running} {
		t.Run(state, func(t *testing.T) {
			parent := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "test"}}
			parent.Status.Status = state
			server := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: "db-server", Namespace: parent.Namespace, Labels: map[string]string{oceanbaseconst.LabelRefOBCluster: parent.Name}}}
			server.Status.Status = serverstatus.BootstrapReady
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
			defer cancel()
			m := &OBServerManager{Ctx: ctx, OBServer: server, Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(parent).Build()}
			op := genMaintainOBServerAfterBootstrapFlow(m).OperationContext
			op.Task = tWaitOBClusterBootstrapped
			op.TaskStatus = taskstatus.Failed
			op.TaskId = "lost"
			op.OnFailure = tt.FailureRule{Strategy: strategy.StartOver, NextTryStatus: serverstatus.Running}
			m.SetOperationContext(op)
			m.HandleFailure()
			if server.Status.Status != serverstatus.BootstrapReady || server.Status.OperationContext != op || op.TaskStatus != taskstatus.Pending || op.TaskId != "" {
				t.Fatal("failed read-only wait skipped the remaining database registration steps")
			}
			err := WaitOBClusterBootstrapped(m)
			if state == clusterstatus.New {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("uninitialized parent returned %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
