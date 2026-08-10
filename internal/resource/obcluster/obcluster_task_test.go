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

package obcluster

import (
	"context"
	"strings"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	logserviceclusterstatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicecluster"
)

type stagedLogServiceClient struct {
	client.Client
	calls         int
	readyAfter    int
	notFoundFirst bool
}

func (c *stagedLogServiceClient) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	c.calls++
	if c.notFoundFirst && c.calls == 1 {
		return apierrors.NewNotFound(schema.GroupResource{Group: v1alpha1.GroupVersion.Group, Resource: "oblogserviceclusters"}, key.Name)
	}
	lsCluster := obj.(*v1alpha1.OBLogServiceCluster)
	lsCluster.Name = key.Name
	if c.readyAfter > 0 && c.calls >= c.readyAfter {
		lsCluster.Status.Status = logserviceclusterstatus.Running
	} else {
		lsCluster.Status.Status = logserviceclusterstatus.New
	}
	return nil
}

func TestWaitForLogServiceReadyRetriesUntilRunning(t *testing.T) {
	kubeClient := &stagedLogServiceClient{readyAfter: 3, notFoundFirst: true}
	err := waitForLogServiceReady(
		context.Background(),
		kubeClient,
		types.NamespacedName{Namespace: "default", Name: "logservice"},
		100*time.Millisecond,
		time.Millisecond,
	)
	if err != nil {
		t.Fatalf("waitForLogServiceReady() returned an unexpected error: %v", err)
	}
	if kubeClient.calls < 3 {
		t.Fatalf("waitForLogServiceReady() made %d calls, want at least 3", kubeClient.calls)
	}
}

func TestWaitForLogServiceReadyReportsLastStatusOnTimeout(t *testing.T) {
	kubeClient := &stagedLogServiceClient{}
	err := waitForLogServiceReady(
		context.Background(),
		kubeClient,
		types.NamespacedName{Namespace: "default", Name: "logservice"},
		10*time.Millisecond,
		time.Millisecond,
	)
	if err == nil {
		t.Fatal("waitForLogServiceReady() returned nil, want a timeout error")
	}
	if !strings.Contains(err.Error(), "last observed status: new") {
		t.Fatalf("waitForLogServiceReady() error = %q, want last observed status", err)
	}
}
