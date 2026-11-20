/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package inspection

import (
	"context"
	"fmt"
	"strings"
	"testing"

	insconst "github.com/oceanbase/ob-operator/internal/dashboard/business/inspection/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/config"
	insmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/inspection"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	batchv1 "k8s.io/api/batch/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

func inspectionMeta() *response.OBClusterMetaBasic {
	return &response.OBClusterMetaBasic{Name: "test-cluster", Namespace: "db-test", UID: "cluster-uid", ClusterName: "database"}
}

func TestInspectionRoleConfigCompatibility(t *testing.T) {
	for _, name := range []string{"", "dashboard-qa-cluster-role"} {
		cfg := &config.Config{Inspection: config.InspectionConfig{ClusterRoleName: name}}
		want := name
		if want == "" {
			want = insconst.ClusterRoleName
		}
		if got := inspectionClusterRoleName(cfg); got != want {
			t.Fatalf("role = %q, want %q", got, want)
		}
	}
}

func TestInspectionCronJobUsesConfiguredRole(t *testing.T) {
	for _, roleName := range []string{"oceanbase-dashboard-cluster-role", "dashboard-qa-cluster-role"} {
		t.Run(roleName, func(t *testing.T) {
			kube := fake.NewSimpleClientset(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: roleName}})
			cfg := &config.Config{Inspection: config.InspectionConfig{ClusterRoleName: roleName, OBHelper: config.ToolConfig{Image: "helper:test"}, OBDiag: config.ToolConfig{Image: "obdiag:test"}}}
			meta := inspectionMeta()
			for _, scenario := range []insmodel.InspectionScenario{insmodel.ScenarioBasic, insmodel.ScenarioPerformance} {
				if err := createCronJobForInspectionWithClient(context.Background(), kube, cfg, meta, &insmodel.InspectionScheduleConfig{Scenario: scenario, Schedule: "59 23 * * *"}); err != nil {
					t.Fatal(err)
				}
			}
			binding, err := kube.RbacV1().ClusterRoleBindings().Get(context.Background(), "ob-ins-db-test-test-cluster", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if binding.RoleRef.Name != roleName || len(binding.Subjects) != 1 || binding.Subjects[0].Namespace != meta.Namespace || binding.Subjects[0].Name != "ob-ins-test-cluster" {
				t.Fatalf("unexpected binding: %+v", binding)
			}
			jobs, err := kube.BatchV1().CronJobs(meta.Namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil || len(jobs.Items) != 2 {
				t.Fatalf("cronjobs=%v error=%v", jobs, err)
			}
			for _, job := range jobs.Items {
				if job.Spec.Schedule != "59 23 * * *" || job.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName != "ob-ins-test-cluster" || job.Spec.TimeZone == nil || *job.Spec.TimeZone != "Asia/Shanghai" {
					t.Fatalf("unexpected cronjob: %+v", job)
				}
				if job.OwnerReferences[0].UID != "cluster-uid" {
					t.Fatal("lost cluster owner reference")
				}
			}
		})
	}
}

func TestInspectionMissingRoleCreatesNoResources(t *testing.T) {
	kube := fake.NewSimpleClientset()
	err := createCronJobForInspectionWithClient(context.Background(), kube, &config.Config{Inspection: config.InspectionConfig{ClusterRoleName: "missing-role"}}, inspectionMeta(), &insmodel.InspectionScheduleConfig{Scenario: insmodel.ScenarioBasic, Schedule: "59 23 * * *"})
	if err == nil || !strings.Contains(err.Error(), "missing-role") {
		t.Fatalf("error = %v", err)
	}
	for _, action := range kube.Actions() {
		if action.GetVerb() != "get" {
			t.Fatalf("unexpected mutation: %+v", action)
		}
	}
}

func TestInspectionAccessRepairsLegacyBinding(t *testing.T) {
	meta := inspectionMeta()
	old := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "ob-ins-db-test-test-cluster", UID: "binding-uid", ResourceVersion: "12"}, Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "ob-ins-test-cluster", Namespace: meta.Namespace}}, RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", APIGroup: rbacv1.GroupName, Name: insconst.ClusterRoleName}}
	kube := fake.NewSimpleClientset(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "dashboard-qa-cluster-role"}}, old)
	if err := ensureInspectionAccess(context.Background(), kube, meta, "dashboard-qa-cluster-role"); err != nil {
		t.Fatal(err)
	}
	got, err := kube.RbacV1().ClusterRoleBindings().Get(context.Background(), old.Name, metav1.GetOptions{})
	if err != nil || got.RoleRef.Name != "dashboard-qa-cluster-role" {
		t.Fatalf("binding=%v error=%v", got, err)
	}
	var deletes int
	for _, action := range kube.Actions() {
		if action.GetVerb() == "delete" {
			deletes++
			pre := action.(kubetesting.DeleteAction).GetDeleteOptions().Preconditions
			if pre == nil || pre.UID == nil || *pre.UID != old.UID || pre.ResourceVersion == nil || *pre.ResourceVersion != old.ResourceVersion {
				t.Fatal("missing replacement preconditions")
			}
		}
	}
	if deletes != 1 {
		t.Fatalf("deletes=%d", deletes)
	}
	kube.ClearActions()
	if err := ensureInspectionAccess(context.Background(), kube, meta, "dashboard-qa-cluster-role"); err != nil {
		t.Fatal(err)
	}
	for _, action := range kube.Actions() {
		if action.GetResource().Resource == "clusterrolebindings" && action.GetVerb() != "get" {
			t.Fatal("matching binding was changed")
		}
	}
}

func TestInspectionAccessDoesNotReplaceForeignSubjects(t *testing.T) {
	for _, subject := range []rbacv1.Subject{{Kind: "ServiceAccount", Name: "ob-ins-test-cluster", Namespace: "other"}, {Kind: "User", Name: "unexpected"}} {
		binding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "ob-ins-db-test-test-cluster"}, Subjects: []rbacv1.Subject{subject}, RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", Name: insconst.ClusterRoleName, APIGroup: rbacv1.GroupName}}
		kube := fake.NewSimpleClientset(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "dashboard-qa-cluster-role"}}, binding)
		if err := ensureInspectionAccess(context.Background(), kube, inspectionMeta(), "dashboard-qa-cluster-role"); err == nil {
			t.Fatal("accepted foreign subjects")
		}
		for _, action := range kube.Actions() {
			if action.GetResource().Resource == "clusterrolebindings" && action.GetVerb() != "get" {
				t.Fatal("foreign binding was mutated")
			}
		}
	}
}

func TestInspectionAccessPropagatesAPIErrors(t *testing.T) {
	for _, resource := range []string{"clusterroles", "serviceaccounts", "clusterrolebindings"} {
		t.Run(resource, func(t *testing.T) {
			kube := fake.NewSimpleClientset(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: insconst.ClusterRoleName}})
			kube.PrependReactor("*", resource, func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, k8serrors.NewForbidden(action.GetResource().GroupResource(), "test", fmt.Errorf("denied"))
			})
			if err := ensureInspectionAccess(context.Background(), kube, inspectionMeta(), insconst.ClusterRoleName); err == nil || !strings.Contains(err.Error(), "denied") {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestInspectionUpdateRepairsAccessBeforeSavingSchedule(t *testing.T) {
	meta := inspectionMeta()
	job := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: "existing", Namespace: meta.Namespace}, Spec: batchv1.CronJobSpec{Schedule: "0 0 * * *"}}
	old := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "ob-ins-db-test-test-cluster"}, Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "ob-ins-test-cluster", Namespace: meta.Namespace}}, RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", APIGroup: rbacv1.GroupName, Name: insconst.ClusterRoleName}}
	kube := fake.NewSimpleClientset(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "dashboard-qa-cluster-role"}}, old, job)
	cfg := &config.Config{Inspection: config.InspectionConfig{ClusterRoleName: "dashboard-qa-cluster-role"}}
	if err := updateCronJobForInspectionWithClient(context.Background(), kube, cfg, job, meta, false, &insmodel.InspectionScheduleConfig{Schedule: "59 23 * * *"}); err != nil {
		t.Fatal(err)
	}
	got, err := kube.BatchV1().CronJobs(meta.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
	if err != nil || got.Spec.Schedule != "59 23 * * *" || got.Spec.Suspend == nil || *got.Spec.Suspend {
		t.Fatalf("job=%v error=%v", got, err)
	}
	binding, err := kube.RbacV1().ClusterRoleBindings().Get(context.Background(), old.Name, metav1.GetOptions{})
	if err != nil || binding.RoleRef.Name != cfg.Inspection.ClusterRoleName {
		t.Fatalf("binding=%v error=%v", binding, err)
	}
}

func TestInspectionUpdateMissingRolePreservesSchedule(t *testing.T) {
	job := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: "existing", Namespace: "db-test"}, Spec: batchv1.CronJobSpec{Schedule: "0 0 * * *"}}
	kube := fake.NewSimpleClientset(job)
	err := updateCronJobForInspectionWithClient(context.Background(), kube, &config.Config{}, job, inspectionMeta(), false, &insmodel.InspectionScheduleConfig{Schedule: "59 23 * * *"})
	if err == nil || job.Spec.Schedule != "0 0 * * *" || job.Spec.Suspend != nil {
		t.Fatalf("error=%v job=%v", err, job)
	}
	for _, action := range kube.Actions() {
		if action.GetVerb() != "get" {
			t.Fatalf("unexpected mutation: %+v", action)
		}
	}
}

func TestInspectionReplacementConflictDoesNotRecreateBinding(t *testing.T) {
	meta := inspectionMeta()
	old := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "ob-ins-db-test-test-cluster"}, Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "ob-ins-test-cluster", Namespace: meta.Namespace}}, RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", Name: insconst.ClusterRoleName, APIGroup: rbacv1.GroupName}}
	kube := fake.NewSimpleClientset(&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "dashboard-qa-cluster-role"}}, old)
	kube.PrependReactor("delete", "clusterrolebindings", func(action kubetesting.Action) (bool, runtime.Object, error) {
		return true, nil, k8serrors.NewConflict(action.GetResource().GroupResource(), old.Name, fmt.Errorf("concurrent change"))
	})
	if err := ensureInspectionAccess(context.Background(), kube, meta, "dashboard-qa-cluster-role"); err == nil || !strings.Contains(err.Error(), "concurrent change") {
		t.Fatalf("error=%v", err)
	}
	for _, action := range kube.Actions() {
		if action.GetResource().Resource == "clusterrolebindings" && action.GetVerb() == "create" {
			t.Fatal("recreated after failed delete")
		}
	}
}
