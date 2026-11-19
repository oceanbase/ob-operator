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
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
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
