package obzone

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
)

func rollingFixture(t *testing.T, oldNames []string, ready, hold bool) (*OBZoneManager, *int) {
	t.Helper()
	cfg := obcfg.GetConfig()
	oldDelete, oldWait, oldDisabled := cfg.Time.ServerDeleteTimeoutSeconds, cfg.Time.DefaultStateWaitTimeout, cfg.Telemetry.Disabled
	cfg.Time.ServerDeleteTimeoutSeconds, cfg.Time.DefaultStateWaitTimeout, cfg.Telemetry.Disabled = 1, 1, true
	t.Cleanup(func() {
		cfg.Time.ServerDeleteTimeoutSeconds, cfg.Time.DefaultStateWaitTimeout, cfg.Telemetry.Disabled = oldDelete, oldWait, oldDisabled
	})
	template := func(sc string) *apitypes.OBServerTemplate {
		storage := &apitypes.StorageSpec{StorageClass: sc}
		return &apitypes.OBServerTemplate{Storage: &apitypes.OceanbaseStorageSpec{DataStorage: storage, LogStorage: storage}}
	}
	zone := &v1alpha1.OBZone{TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBZone"},
		ObjectMeta: metav1.ObjectMeta{Name: "zone1", Namespace: "test", UID: "zone-uid"},
		Spec:       v1alpha1.OBZoneSpec{Topology: apitypes.OBZoneTopology{Zone: "zone1", Replica: len(oldNames)}, OBServerTemplate: template("new")}}
	var objects []client.Object
	for _, name := range oldNames {
		old := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: zone.Namespace, UID: types.UID("uid-" + name),
			OwnerReferences: []metav1.OwnerReference{{Kind: "OBZone", APIVersion: zone.APIVersion, Name: zone.Name, UID: zone.UID}},
			Labels:          map[string]string{oceanbaseconst.LabelRefOBZone: zone.Name}},
			Spec: v1alpha1.OBServerSpec{OBServerTemplate: template("old")}}
		if hold {
			old.Finalizers = []string{"test.example/hold"}
		}
		objects = append(objects, old)
	}
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	created := 0
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).
		WithInterceptorFuncs(interceptor.Funcs{Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			obj.SetUID(types.UID("uid-" + obj.GetName()))
			if ready {
				obj.(*v1alpha1.OBServer).Status.Status = serverstatus.Running
			}
			err := cl.Create(ctx, obj, opts...)
			if err == nil {
				created++
			}
			return err
		}}).Build()
	ctx := context.Background()
	logger := logr.Discard()
	return &OBZoneManager{Ctx: ctx, OBZone: zone, Client: cl, Logger: &logger, Recorder: telemetry.NewRecorder(ctx, record.NewFakeRecorder(20))}, &created
}

func TestRollingStorageWaitsForPreviousObserverDeletion(t *testing.T) {
	m, count := rollingFixture(t, []string{"old-1", "old-2"}, true, true)
	for attempt := 1; attempt <= 2; attempt++ {
		err := RollingReplaceOBServers(m)
		if err == nil || !strings.Contains(err.Error(), "old-1") || *count != 1 {
			t.Fatalf("attempt %d, deletion held by finalizer: got %v, created %d replacements; want error identifying old-1 and one replacement", attempt, err, *count)
		}
	}
	// Once both old finalizers release, the same task should keep the first
	// successful replacement and migrate the second server exactly once.
	for _, name := range []string{"old-1", "old-2"} {
		server := &v1alpha1.OBServer{}
		if err := m.Client.Get(m.Ctx, client.ObjectKey{Namespace: m.OBZone.Namespace, Name: name}, server); err != nil {
			t.Fatal(err)
		}
		server.Finalizers = nil
		if err := m.Client.Update(m.Ctx, server); err != nil {
			t.Fatal(err)
		}
	}
	if err := RollingReplaceOBServers(m); err != nil || *count != 2 {
		t.Fatalf("after releasing the finalizers: got %v and %d replacements, want success and exactly two", err, *count)
	}
	current, err := m.listOBServers()
	if err != nil || len(current.Items) != 2 {
		t.Fatalf("final servers: got %+v, error %v", current, err)
	}
	for _, server := range current.Items {
		if m.checkIfStorageClassChanged(&server) {
			t.Fatalf("old storage still present on %s", server.Name)
		}
	}
}

func TestRollingStorageResumesUnreadyReplacement(t *testing.T) {
	for _, scenario := range []string{"restart", "legacy", "changed owner", "changed source"} {
		t.Run(scenario, func(t *testing.T) {
			m, count := rollingFixture(t, []string{"old"}, false, false)
			if err := RollingReplaceOBServers(m); err == nil || *count != 1 {
				t.Fatalf("want one pending replacement, got %d, err %v", *count, err)
			}
			list, err := m.listOBServers()
			if err != nil || len(list.Items) != 2 {
				t.Fatalf("before restart: %+v, %v", list, err)
			}
			var target *v1alpha1.OBServer
			for i := range list.Items {
				if list.Items[i].Name != "old" {
					target = &list.Items[i]
				}
			}
			if target == nil {
				t.Fatal("old server lost before readiness")
			}
			// A newly constructed manager has no memory of the prior task.
			restarted := *m
			if err := RollingReplaceOBServers(&restarted); err == nil || *count != 1 {
				t.Fatalf("retry created %d replacements, err %v", *count, err)
			}
			const sourceAnnotation = "oceanbase.oceanbase.com/replaces-observer-uid"
			target.Status.Status = serverstatus.Running
			switch scenario {
			case "legacy":
				delete(target.Annotations, sourceAnnotation)
			case "changed owner":
				target.OwnerReferences[0].UID = "other-zone"
			case "changed source":
				target.Annotations[sourceAnnotation] = "other-source"
			}
			if err := m.Client.Update(m.Ctx, target); err != nil {
				t.Fatal(err)
			}
			err = RollingReplaceOBServers(&restarted)
			after, listErr := m.listOBServers()
			if listErr != nil {
				t.Fatal(listErr)
			}
			if strings.HasPrefix(scenario, "changed") {
				if err == nil || len(after.Items) != 2 || *count != 1 {
					t.Fatalf("unsafe adoption: %d live servers, %d creates, error %v", len(after.Items), *count, err)
				}
				return
			}
			if err != nil || len(after.Items) != 1 || after.Items[0].UID != target.UID || *count != 1 {
				t.Fatalf("expected original replacement: live %+v, creates %d, error %v", after.Items, *count, err)
			}
		})
	}
}

func TestRollingStorageFlowRetriesCurrentTask(t *testing.T) {
	m, _ := rollingFixture(t, []string{"old"}, false, false)
	m.OBZone.Status.Status = zonestatus.RollingUpdateServers
	f, err := m.GetTaskFlow()
	if err != nil || f.OperationContext.OnFailure.Strategy != strategy.RetryFromCurrent {
		t.Fatalf("new storage flow: %+v, %v", f, err)
	}
	m.OBZone.Status.OperationContext = f.OperationContext.DeepCopy()
	if !m.CanReplayTask(tRollingReplaceOBServers) {
		t.Fatal("lost rolling replacement should resume using its persisted identity")
	}
	m.OBZone.Status.OperationContext.OnFailure.Strategy = ""
	f, err = m.GetTaskFlow()
	if err != nil || f.OperationContext.OnFailure.Strategy != strategy.RetryFromCurrent {
		t.Fatalf("unpopulated previous-version policy: %+v, %v", f, err)
	}
	m.OBZone.Status.OperationContext.OnFailure.Strategy = strategy.StartOver
	m.OBZone.Status.OperationContext.OnFailure.RetryCount = 3
	f, err = m.GetTaskFlow()
	if err != nil || f.OperationContext.OnFailure.Strategy != strategy.RetryFromCurrent || f.OperationContext.OnFailure.RetryCount != 3 {
		t.Fatalf("persisted previous-version storage flow: %+v, %v", f, err)
	}
}
