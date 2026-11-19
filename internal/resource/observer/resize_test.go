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
	"testing"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
)

func TestObserverPVCExpansion(t *testing.T) {
	pvc := func(name, requested, capacity string) corev1.PersistentVolumeClaim {
		return corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       corev1.PersistentVolumeClaimSpec{Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(requested)}}},
			Status:     corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound, Capacity: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(capacity)}},
		}
	}
	data := pvc("s-data-file", "61Gi", "61Gi")
	log := pvc("s-log", "20Gi", "20Gi")
	redo := pvc("s-data-log", "60Gi", "60Gi")
	unbound := data.DeepCopy()
	unbound.Status.Phase = corev1.ClaimPending
	deleting := data.DeepCopy()
	now := metav1.Now()
	deleting.DeletionTimestamp = &now
	for _, tc := range []struct {
		name                     string
		single, shared, omitRedo bool
		pvcs                     []corev1.PersistentVolumeClaim
		pending                  bool
	}{
		{name: "request was written, storage still small", pvcs: []corev1.PersistentVolumeClaim{pvc(data.Name, "61Gi", "60Gi"), log, redo}, pending: true},
		{name: "existing request smaller", pvcs: []corev1.PersistentVolumeClaim{pvc(data.Name, "60Gi", "62Gi"), log, redo}, pending: true},
		{name: "resized", pvcs: []corev1.PersistentVolumeClaim{data, log, redo}},
		{name: "driver rounds up", pvcs: []corev1.PersistentVolumeClaim{pvc(data.Name, "61Gi", "62Gi"), log, redo}},
		{name: "claims missing", pending: true},
		{name: "redo missing", pvcs: []corev1.PersistentVolumeClaim{data, log}, pending: true},
		{name: "optional redo absent", omitRedo: true, pvcs: []corev1.PersistentVolumeClaim{data, log}},
		{name: "unbound", pvcs: []corev1.PersistentVolumeClaim{*unbound, log, redo}, pending: true},
		{name: "being deleted", pvcs: []corev1.PersistentVolumeClaim{*deleting, log, redo}, pending: true},
		{name: "single sum", single: true, pvcs: []corev1.PersistentVolumeClaim{pvc("s", "141Gi", "141Gi")}},
		{name: "single lag", single: true, pvcs: []corev1.PersistentVolumeClaim{pvc("s", "141Gi", "140Gi")}, pending: true},
		{name: "previous single claim", pvcs: []corev1.PersistentVolumeClaim{pvc("s", "141Gi", "141Gi")}},
		{name: "shared storage excludes redo", shared: true, pvcs: []corev1.PersistentVolumeClaim{data, log}},
		{name: "shared single excludes redo", shared: true, single: true, pvcs: []corev1.PersistentVolumeClaim{pvc("s", "81Gi", "81Gi")}},
		{name: "unrelated PVC", pvcs: []corev1.PersistentVolumeClaim{pvc("other-data-file", "61Gi", "61Gi"), redo, log}, pending: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &OBServerManager{OBServer: &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: "s", Annotations: map[string]string{}}}}
			storage := &apitypes.OceanbaseStorageSpec{DataStorage: &apitypes.StorageSpec{Size: resource.MustParse("61Gi")}, LogStorage: &apitypes.StorageSpec{Size: resource.MustParse("20Gi")}}
			if !tc.omitRedo {
				storage.RedoLogStorage = &apitypes.StorageSpec{Size: resource.MustParse("60Gi")}
			}
			m.OBServer.Spec.OBServerTemplate = &apitypes.OBServerTemplate{Storage: storage}
			if tc.single {
				m.OBServer.Annotations[oceanbaseconst.AnnotationsSinglePVC] = "true"
			}
			if tc.shared {
				m.OBServer.Annotations[oceanbaseconst.AnnotationsDeploymentMode] = oceanbaseconst.DeploymentModeSharedStorage
			}
			if got := m.checkIfStorageExpand(&corev1.PersistentVolumeClaimList{Items: tc.pvcs}); got != tc.pending {
				t.Fatalf("expansion pending = %t, want %t", got, tc.pending)
			}
		})
	}
}

func TestLegacyObserverPVCExpansion(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		claimOwner, podOwner, zoneLabel string
		mounted, absent, wantUpdate     bool
	}{
		{name: "legacy owned claim", claimOwner: "server-uid", zoneLabel: "zone-uid", wantUpdate: true},
		{name: "legacy independent mounted claim", podOwner: "server-uid", zoneLabel: "zone-uid", mounted: true, wantUpdate: true},
		{name: "unlabelled independent mounted claim", podOwner: "server-uid", mounted: true, wantUpdate: true},
		{name: "foreign claim owner", claimOwner: "old-server", podOwner: "server-uid", zoneLabel: "zone-uid", mounted: true},
		{name: "foreign pod owner", podOwner: "other-server", zoneLabel: "zone-uid", mounted: true},
		{name: "unrelated volume", podOwner: "server-uid", zoneLabel: "zone-uid"},
		{name: "foreign ref uid", podOwner: "server-uid", zoneLabel: "other-zone", mounted: true},
		{name: "missing volume", podOwner: "server-uid", mounted: true, absent: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "ns", UID: "server-uid", Labels: map[string]string{oceanbaseconst.LabelRefUID: "zone-uid"}, Annotations: map[string]string{oceanbaseconst.AnnotationsSinglePVC: "true"}}}
			srv.Spec.OBServerTemplate = &apitypes.OBServerTemplate{Storage: &apitypes.OceanbaseStorageSpec{DataStorage: &apitypes.StorageSpec{Size: resource.MustParse("61Gi")}, LogStorage: &apitypes.StorageSpec{Size: resource.MustParse("20Gi")}}}
			ref := func(uid string) []metav1.OwnerReference {
				if uid == "" {
					return nil
				}
				return []metav1.OwnerReference{{Kind: "OBServer", Name: "s", UID: types.UID(uid)}}
			}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "ns", OwnerReferences: ref(tc.podOwner)}}
			if tc.mounted {
				pod.Spec.Volumes = []corev1.Volume{{Name: "data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "s"}}}}
			}
			pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "ns", OwnerReferences: ref(tc.claimOwner), Labels: map[string]string{oceanbaseconst.LabelRefUID: tc.zoneLabel}}, Spec: corev1.PersistentVolumeClaimSpec{Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("80Gi")}}}, Status: corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound, Capacity: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("80Gi")}}}
			builder := fake.NewClientBuilder().WithObjects(pod)
			if !tc.absent {
				builder.WithObjects(pvc)
			}
			m := &OBServerManager{Ctx: context.Background(), OBServer: srv, K8sResClient: builder.Build()}
			err := ExpandPVC(m)
			if (err == nil) != tc.wantUpdate {
				t.Fatalf("unexpected expansion result: %v", err)
			}
			if tc.absent {
				return
			}
			var stored corev1.PersistentVolumeClaim
			if err := m.K8sResClient.Get(m.Ctx, client.ObjectKeyFromObject(pvc), &stored); err != nil {
				t.Fatal(err)
			}
			want := "80Gi"
			if tc.wantUpdate {
				want = "81Gi"
			}
			if stored.Spec.Resources.Requests.Storage().Cmp(resource.MustParse(want)) != 0 || stored.Labels[oceanbaseconst.LabelRefUID] != tc.zoneLabel {
				t.Fatalf("claim changed unexpectedly: %+v", stored)
			}
		})
	}
}

func TestPodRecoveryDuringPVCExpansion(t *testing.T) {
	for _, tc := range []struct{ name, mode, want string }{
		{"service-address", "service", serverstatus.Recover},
		{"dynamic-address", "", serverstatus.Unrecoverable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			server := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{Name: "s", Namespace: "ns", UID: "uid", Annotations: map[string]string{oceanbaseconst.AnnotationsMode: tc.mode}}}
			server.Status.Status = serverstatus.ExpandPVC
			logger := logr.Discard()
			m := &OBServerManager{Ctx: context.Background(), OBServer: server, Logger: &logger,
				K8sResClient: fake.NewClientBuilder().Build(), Client: fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(server).WithObjects(server.DeepCopy()).Build()}
			if err := WaitForPvcResized(m); err != nil {
				t.Fatal(err)
			}
			m.SetOperationContext(genExpandPVCFlow(m).OperationContext)
			m.FinishTask()
			if err := m.UpdateStatus(); err != nil {
				t.Fatal(err)
			}
			if server.Status.Status != tc.want {
				t.Fatalf("status = %s, want %s", server.Status.Status, tc.want)
			}
		})
	}
	apiErr := errors.New("temporarily unavailable")
	resourceClient := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
		return apiErr
	}}).Build()
	manager := &OBServerManager{OBServer: &v1alpha1.OBServer{}, Ctx: context.Background(), K8sResClient: resourceClient}
	if err := WaitForPvcResized(manager); !errors.Is(err, apiErr) {
		t.Fatalf("unexpected result for temporary API error: %v", err)
	}
}
