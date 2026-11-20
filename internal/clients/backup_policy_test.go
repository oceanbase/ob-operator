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

package clients

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func TestFindUnlabelledBackupPolicy(t *testing.T) {
	tenant := &v1alpha1.OBTenant{ObjectMeta: metav1.ObjectMeta{Name: "tenant-cr", Namespace: "qa"},
		Spec: v1alpha1.OBTenantSpec{ClusterName: "cluster", TenantName: "db_tenant"}}
	policy := v1alpha1.OBTenantBackupPolicy{ObjectMeta: metav1.ObjectMeta{Name: "nightly", Namespace: "qa"},
		Spec: v1alpha1.OBTenantBackupPolicySpec{ObClusterName: "cluster", TenantName: "db_tenant"}}
	otherCluster := policy.DeepCopy()
	otherCluster.Spec.ObClusterName = "other"
	otherTenant := policy.DeepCopy()
	otherTenant.Spec.TenantName = "tenant-cr"
	otherCR := policy.DeepCopy()
	otherCR.Spec.TenantCRName = "unrelated"
	labelled := policy.DeepCopy()
	labelled.Labels = map[string]string{oceanbaseconst.LabelTenantName: "unrelated"}
	candidates := []v1alpha1.OBTenantBackupPolicy{*otherCluster, *otherTenant, *otherCR, *labelled}
	got, err := findUnlabelledBackupPolicy(tenant, candidates)
	if err != nil || got != nil {
		t.Fatalf("unrelated policies matched: %v, %v", got, err)
	}
	candidates = append(candidates, policy)
	got, err = findUnlabelledBackupPolicy(tenant, candidates)
	if err != nil || got == nil || got.Name != policy.Name {
		t.Fatalf("database policy not found: %v, %v", got, err)
	}
	other := policy.DeepCopy()
	other.Name = "hourly"
	candidates = append(candidates, *other)
	if got, err := findUnlabelledBackupPolicy(tenant, candidates); err == nil || got != nil {
		t.Fatalf("ambiguous policy match: %v, %v", got, err)
	}
}
