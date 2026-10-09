/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package job

import (
	"strings"
	"testing"

	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestResolveJobStatusReportsFailedCreate(t *testing.T) {
	k8sJob := &batchv1.Job{}
	events := []corev1.Event{{
		Type:    corev1.EventTypeWarning,
		Reason:  "FailedCreate",
		Message: `Error creating: pods "log-test" is forbidden: serviceaccount "dashboard-qa-sa" not found`,
	}}

	status, output := resolveJobStatus(k8sJob, events)
	if status != jobmodel.JobStatusFailed {
		t.Fatalf("resolveJobStatus() status = %q, want %q", status, jobmodel.JobStatusFailed)
	}
	if !strings.Contains(output, "FailedCreate") || !strings.Contains(output, "serviceaccount") {
		t.Fatalf("resolveJobStatus() output = %q, want Kubernetes failure reason", output)
	}
}

func TestResolveJobStatusKeepsPendingWithoutTerminalEvidence(t *testing.T) {
	status, output := resolveJobStatus(&batchv1.Job{}, []corev1.Event{{
		Type:    corev1.EventTypeWarning,
		Reason:  "Backoff",
		Message: "temporary controller retry",
	}})
	if status != jobmodel.JobStatusPending || output != "" {
		t.Fatalf("resolveJobStatus() = (%q, %q), want pending without output", status, output)
	}
}

func TestResolveJobStatusKeepsPendingAfterTransientFailedCreate(t *testing.T) {
	status, output := resolveJobStatus(&batchv1.Job{}, []corev1.Event{{
		Type:    corev1.EventTypeWarning,
		Reason:  "FailedCreate",
		Message: "Internal error occurred: transient API timeout",
	}})
	if status != jobmodel.JobStatusPending || output != "" {
		t.Fatalf("resolveJobStatus() = (%q, %q), want retryable FailedCreate to remain pending", status, output)
	}
}

func TestResolveJobStatusUsesFailedCondition(t *testing.T) {
	k8sJob := &batchv1.Job{Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
		Type:    batchv1.JobFailed,
		Status:  corev1.ConditionTrue,
		Reason:  "BackoffLimitExceeded",
		Message: "Job has reached the specified backoff limit",
	}}}}

	status, output := resolveJobStatus(k8sJob, nil)
	if status != jobmodel.JobStatusFailed {
		t.Fatalf("resolveJobStatus() status = %q, want %q", status, jobmodel.JobStatusFailed)
	}
	if !strings.Contains(output, "BackoffLimitExceeded") {
		t.Fatalf("resolveJobStatus() output = %q, want failed condition reason", output)
	}
}

func TestResolveJobStatusKeepsRunningJobActive(t *testing.T) {
	k8sJob := &batchv1.Job{Status: batchv1.JobStatus{Active: 1}}
	status, output := resolveJobStatus(k8sJob, []corev1.Event{{
		Type:    corev1.EventTypeWarning,
		Reason:  "FailedCreate",
		Message: "an earlier retry failed",
	}})
	if status != jobmodel.JobStatusRunning || output != "" {
		t.Fatalf("resolveJobStatus() = (%q, %q), want running without failure output", status, output)
	}
}
