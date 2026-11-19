/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package obtenantbackuppolicy

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func TestBackupPolicyTenantNameAcrossUpgrade(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	dbTenant := &v1alpha1.OBTenant{Spec: v1alpha1.OBTenantSpec{TenantName: "database_tenant"}}
	for _, tt := range []struct {
		name, explicit string
		status         v1alpha1.OBTenantBackupPolicyStatus
		want           string
	}{
		{"legacy spec", "explicit", v1alpha1.OBTenantBackupPolicyStatus{}, "explicit"},
		{"explicit name overrides cached name", "explicit", v1alpha1.OBTenantBackupPolicyStatus{TenantName: "old"}, "explicit"},
		{"resolved reference", "", v1alpha1.OBTenantBackupPolicyStatus{TenantName: "observed", TenantCR: dbTenant}, "observed"},
		{"deprecated tenant snapshot", "", v1alpha1.OBTenantBackupPolicyStatus{TenantCR: dbTenant}, "database_tenant"},
		{"database record", "", v1alpha1.OBTenantBackupPolicyStatus{TenantInfo: &model.OBTenant{TenantName: "cached"}}, "cached"},
		{"unresolved reference is not a database name", "", v1alpha1.OBTenantBackupPolicyStatus{}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			policy := &v1alpha1.OBTenantBackupPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "qa", Name: "backup"},
				Spec:       v1alpha1.OBTenantBackupPolicySpec{TenantName: tt.explicit, Suspend: true},
				Status:     tt.status,
			}
			if tt.explicit == "" {
				policy.Spec.TenantCRName = "kubernetes-name"
			} else {
				policy.Spec.TenantSecret = "database-secret"
			}
			policy.Status.Status = constants.BackupPolicyStatusPaused
			c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(policy).WithObjects(policy).Build()
			m := &ObTenantBackupPolicyManager{Ctx: context.Background(), Client: c, BackupPolicy: policy}
			if err := m.UpdateStatus(); err != nil {
				t.Fatal(err)
			}
			var stored v1alpha1.OBTenantBackupPolicy
			if err := c.Get(m.Ctx, client.ObjectKeyFromObject(policy), &stored); err != nil {
				t.Fatal(err)
			}
			if stored.Status.TenantName != tt.want {
				t.Errorf("stored tenantName = %q, want %q", stored.Status.TenantName, tt.want)
			}
		})
	}
}
