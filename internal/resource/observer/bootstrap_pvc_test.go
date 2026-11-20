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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func TestBootstrapPVCReplay(t *testing.T) {
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		count       int
		firstSize   string
	}{
		{"separate", nil, 3, "5Gi"},
		{"independent", map[string]string{obconst.AnnotationsIndependentPVCLifecycle: "true"}, 3, "5Gi"},
		{"single", map[string]string{obconst.AnnotationsSinglePVC: "true"}, 1, "10Gi"},
		{"shared", map[string]string{obconst.AnnotationsDeploymentMode: obconst.DeploymentModeSharedStorage}, 2, "5Gi"},
		{"single shared", map[string]string{obconst.AnnotationsSinglePVC: "true", obconst.AnnotationsDeploymentMode: obconst.DeploymentModeSharedStorage}, 1, "8Gi"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := &v1alpha1.OBServer{ObjectMeta: metav1.ObjectMeta{
				Name: "s", Namespace: "ns", UID: "server-uid", Annotations: tc.annotations,
				Labels: map[string]string{obconst.LabelRefUID: "zone-uid"},
			}}
			spec := func(size string) *apitypes.StorageSpec {
				return &apitypes.StorageSpec{StorageClass: "local", Size: resource.MustParse(size)}
			}
			server.Spec.OBServerTemplate = &apitypes.OBServerTemplate{Storage: &apitypes.OceanbaseStorageSpec{
				DataStorage: spec("5Gi"), LogStorage: spec("3Gi"), RedoLogStorage: spec("2Gi"),
			}}
			first := "s-" + obconst.DataVolumeSuffix
			if tc.count == 1 {
				first = "s"
			}
			claim := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: first, Namespace: "ns", UID: "original-pvc", ResourceVersion: "42",
					Labels: map[string]string{obconst.LabelRefUID: "server-uid", obconst.LabelRefOBServer: "s"},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: ptr.To("local"), VolumeName: "bound-pv",
					VolumeMode: ptr.To(corev1.PersistentVolumeFilesystem), AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(tc.firstSize)}},
				},
				Status: corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound},
			}
			if tc.annotations[obconst.AnnotationsIndependentPVCLifecycle] != "true" {
				claim.OwnerReferences = []metav1.OwnerReference{{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBServer", Name: "s", UID: "server-uid"}}
			}
			pv := &corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "bound-pv"}}
			pv.Spec.ClaimRef = &corev1.ObjectReference{Name: first, Namespace: "ns", UID: claim.UID}
			cl := fake.NewClientBuilder().WithObjects(claim, pv).Build()
			stale := interceptor.NewClient(cl, interceptor.Funcs{
				Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
					return errors.New("claim has not reached the informer cache")
				},
			})
			m := &OBServerManager{OBServer: server, Ctx: context.Background(), K8sResClient: stale, K8sReader: cl}
			for range 2 {
				if err := CreateOBServerPVC(m); err != nil {
					t.Fatal(err)
				}
			}
			var claims corev1.PersistentVolumeClaimList
			if err := cl.List(m.Ctx, &claims); err != nil {
				t.Fatal(err)
			}
			if len(claims.Items) != tc.count {
				t.Fatalf("found %d claims, want %d", len(claims.Items), tc.count)
			}
			var got corev1.PersistentVolumeClaim
			if err := cl.Get(m.Ctx, client.ObjectKeyFromObject(claim), &got); err != nil {
				t.Fatal(err)
			}
			if got.UID != claim.UID || got.Spec.VolumeName != "bound-pv" || got.ResourceVersion != claim.ResourceVersion || server.Labels[obconst.LabelRefUID] != "zone-uid" {
				t.Fatal("replay modified the existing claim or parent identity")
			}
			pv.Spec.ClaimRef.UID = "previous-pvc-uid"
			if err := cl.Update(m.Ctx, pv); err != nil {
				t.Fatal(err)
			}
			if err := CreateOBServerPVC(m); err == nil {
				t.Fatal("replayed a binding to another PVC")
			}
			got.Labels[obconst.LabelRefUID] = "another-observer"
			if err := cl.Update(m.Ctx, &got); err != nil {
				t.Fatal(err)
			}
			if err := CreateOBServerPVC(m); err == nil {
				t.Fatal("replayed mismatched claim")
			}
		})
	}
}

func TestBootstrapPVCCollisions(t *testing.T) {
	want := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          map[string]string{obconst.LabelRefUID: "uid", obconst.LabelRefOBServer: "s"},
			OwnerReferences: []metav1.OwnerReference{{Kind: "OBServer", Name: "s", UID: "uid"}},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: ptr.To("local"),
			Resources:        corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("2Gi")}},
		},
	}
	legacy := want.DeepCopy()
	legacy.Labels = map[string]string{obconst.LabelRefUID: "zone-uid"}
	if !matchesBootstrapClaim(want, legacy, "zone-uid") {
		t.Fatal("failed to verify an older claim with this Observer owner and its Zone label")
	}
	independent := want.DeepCopy()
	independent.OwnerReferences = nil
	legacy.OwnerReferences = nil
	if matchesBootstrapClaim(independent, legacy, "zone-uid") {
		t.Fatal("accepted an unowned historical claim with only a Zone identity")
	}
	for name, change := range map[string]func(*corev1.PersistentVolumeClaim){
		"previous observer":   func(p *corev1.PersistentVolumeClaim) { p.OwnerReferences[0].UID = "old" },
		"missing owner":       func(p *corev1.PersistentVolumeClaim) { p.OwnerReferences = nil },
		"missing label":       func(p *corev1.PersistentVolumeClaim) { delete(p.Labels, obconst.LabelRefOBServer) },
		"other storage class": func(p *corev1.PersistentVolumeClaim) { p.Spec.StorageClassName = ptr.To("another") },
		"storage request": func(p *corev1.PersistentVolumeClaim) {
			p.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("1Gi")
		},
		"block":    func(p *corev1.PersistentVolumeClaim) { p.Spec.VolumeMode = ptr.To(corev1.PersistentVolumeBlock) },
		"deleting": func(p *corev1.PersistentVolumeClaim) { p.DeletionTimestamp = ptr.To(metav1.Now()) },
		"lost":     func(p *corev1.PersistentVolumeClaim) { p.Status.Phase = corev1.ClaimLost },
	} {
		t.Run(name, func(t *testing.T) {
			p := want.DeepCopy()
			change(p)
			if matchesBootstrapClaim(want, p, "zone-uid") {
				t.Fatal("accepted mismatched claim")
			}
		})
	}
}
