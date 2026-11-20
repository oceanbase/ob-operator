package obcluster

import (
	"context"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type testBootstrapStatement struct {
	send   func() error
	closed atomic.Bool
}

func (s *testBootstrapStatement) Exec(context.Context) error { return s.send() }
func (s *testBootstrapStatement) Close()                     { s.closed.Store(true) }

func dispatchFixture(t *testing.T, edit func(*v1alpha1.OBCluster)) *OBClusterManager {
	t.Helper()
	cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{Namespace: "db", Name: "cluster", UID: "original", Generation: 4}}
	cluster.Status.Status = clusterstatus.New
	cluster.Status.OperationContext = &tasktypes.OperationContext{Name: "bootstrap obcluster", Task: tDispatchClassicBootstrap, TaskStatus: taskstatus.Running}
	live := cluster.DeepCopy()
	if edit != nil {
		edit(live)
	}
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	// v0.16's in-memory tracker separates its resourceVersion comparison from
	// its store update. Serialize only API writes to model Kubernetes' atomic
	// write transaction; the workers' independent reads still race.
	var writes sync.Mutex
	kube := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(cluster).WithObjects(live).
		WithInterceptorFuncs(interceptor.Funcs{Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, options ...client.PatchOption) error {
			writes.Lock()
			defer writes.Unlock()
			return c.Patch(ctx, obj, patch, options...)
		}}).Build()
	return &OBClusterManager{Ctx: context.Background(), OBCluster: cluster, Client: kube, APIReader: kube}
}

func TestClassicBootstrapRecoversBeforeDispatchAndClaimsOnce(t *testing.T) {
	m := dispatchFixture(t, nil)
	notSent := errors.New("preparation interrupted")
	var writes, reads atomic.Int32
	observe := func() error { reads.Add(1); return nil }
	if err := m.dispatchClassicBootstrap(func() (bootstrapStatement, error) { return nil, notSent }, observe); !errors.Is(err, notSent) {
		t.Fatalf("lost preflight: %v", err)
	}
	if _, claimed, err := m.currentClassicBootstrap(m.Ctx); err != nil || claimed || reads.Load() != 0 {
		t.Fatalf("preflight left a claim or attempted observation: %v %v", claimed, err)
	}
	// Force both workers to submit patches based on the same live version.
	var atPatch sync.WaitGroup
	atPatch.Add(2)
	var attempts, conflicts atomic.Int32
	m.Client = interceptor.NewClient(m.Client.(client.WithWatch), interceptor.Funcs{Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, options ...client.PatchOption) error {
		if attempts.Add(1) <= 2 {
			atPatch.Done()
			atPatch.Wait()
		}
		err := c.Patch(ctx, obj, patch, options...)
		if apierrors.IsConflict(err) {
			conflicts.Add(1)
		}
		return err
	}})
	prepare := func() (bootstrapStatement, error) {
		return &testBootstrapStatement{send: func() error {
			if _, claimed, err := m.currentClassicBootstrap(m.Ctx); err != nil || !claimed {
				return errors.New("SQL invoked before durable claim")
			}
			writes.Add(1)
			return nil
		}}, nil
	}
	results := make(chan error, 2)
	for range 2 {
		go func() { results <- m.dispatchClassicBootstrap(prepare, observe) }()
	}
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if writes.Load() != 1 || reads.Load() != 1 || conflicts.Load() != 1 {
		t.Fatalf("concurrent prepared workers issued %d statements, %d verifications, %d API conflicts", writes.Load(), reads.Load(), conflicts.Load())
	}
	if err := m.dispatchClassicBootstrap(func() (bootstrapStatement, error) {
		return nil, errors.New("a claimed task prepared SQL again")
	}, observe); err != nil || reads.Load() != 2 {
		t.Fatalf("claimed task failed to observe: %v", err)
	}
}

func TestClassicBootstrapValidatesLiveDispatchAndVerifiesAmbiguousSQL(t *testing.T) {
	for _, edit := range []func(*v1alpha1.OBCluster){
		func(c *v1alpha1.OBCluster) { c.UID = "replaced" },
		func(c *v1alpha1.OBCluster) {
			c.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "original/5"}
		},
		func(c *v1alpha1.OBCluster) {
			c.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "original/04"}
		},
		func(c *v1alpha1.OBCluster) {
			c.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "original/0"}
		},
		func(c *v1alpha1.OBCluster) { c.Generation++ },
		func(c *v1alpha1.OBCluster) { c.Status.Status = clusterstatus.Running },
		func(c *v1alpha1.OBCluster) { c.Status.OperationContext.Task = tBootstrap },
		func(c *v1alpha1.OBCluster) { c.Status.OperationContext.TaskStatus = taskstatus.Failed },
		func(c *v1alpha1.OBCluster) { c.Status.OperationContext = nil },
		func(c *v1alpha1.OBCluster) { c.Spec.DeploymentMode = obconst.DeploymentModeSharedStorage },
		func(c *v1alpha1.OBCluster) {
			c.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "another/1"}
		},
		func(c *v1alpha1.OBCluster) {
			now := metav1.Now()
			c.DeletionTimestamp = &now
			c.Finalizers = []string{"keep"}
		},
	} {
		m := dispatchFixture(t, edit)
		if err := m.dispatchClassicBootstrap(func() (bootstrapStatement, error) {
			t.Fatal("invalid cluster or task reached preflight")
			return nil, nil
		}, func() error { t.Fatal("invalid cluster treated as a valid claim"); return nil }); err == nil {
			t.Fatal("invalid dispatch target accepted")
		}
	}
	older := dispatchFixture(t, func(c *v1alpha1.OBCluster) {
		c.Generation++
		c.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "original/4"}
	})
	observed := false
	if err := older.dispatchClassicBootstrap(func() (bootstrapStatement, error) { t.Fatal("old generation claim dispatched again"); return nil, nil }, func() error { observed = true; return nil }); err != nil || !observed {
		t.Fatalf("valid older generation claim failed to observe: %v", err)
	}
	for _, proof := range []error{nil, errors.New("database identity unconfirmed")} {
		m := dispatchFixture(t, nil)
		lost := errors.New("connection lost during SQL")
		stmt := &testBootstrapStatement{send: func() error { return lost }}
		result := m.dispatchClassicBootstrap(func() (bootstrapStatement, error) { return stmt, nil }, func() error {
			if !stmt.closed.Load() {
				t.Fatal("verification holds the SQL connection")
			}
			return proof
		})
		if (proof == nil && result != nil) || (proof != nil && (!errors.Is(result, lost) || !errors.Is(result, proof))) {
			t.Fatalf("SQL error/verification result %v, got %v", proof, result)
		}
	}
}

func TestClassicBootstrapPreservesPersistedAndSharedFlowNames(t *testing.T) {
	m := dispatchFixture(t, nil)
	classic := genBootstrapOBClusterFlow(m).OperationContext.Tasks
	if !slices.Contains(classic, tDispatchClassicBootstrap) || slices.Contains(classic, tBootstrap) ||
		!m.CanReplayTask(tDispatchClassicBootstrap) || m.CanReplayTask(tBootstrap) || m.ObserveLostTask(tBootstrap) == nil {
		t.Fatal("classic flow must distinguish claimed dispatch from an unversioned lost task")
	}
	m.OBCluster.Spec.DeploymentMode = obconst.DeploymentModeSharedStorage
	shared := genBootstrapOBClusterFlow(m).OperationContext.Tasks
	if !slices.Contains(shared, tBootstrap) || slices.Contains(shared, tDispatchClassicBootstrap) || m.ObserveLostTask(tBootstrap) != nil {
		t.Fatal("shared-storage bootstrap changed")
	}
}
