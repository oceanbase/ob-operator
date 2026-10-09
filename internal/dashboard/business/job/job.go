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
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	bizconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	jobmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

func GetJob(ctx context.Context, namespace, name string) (*jobmodel.Job, error) {
	client := k8sclient.GetClient()
	k8sJob, err := client.ClientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var warningEvents []corev1.Event
	if initialJobStatus(k8sJob) == jobmodel.JobStatusPending {
		eventList, err := client.ClientSet.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
			FieldSelector: fields.AndSelectors(
				fields.OneTermEqualSelector("involvedObject.uid", string(k8sJob.UID)),
				fields.OneTermEqualSelector("type", corev1.EventTypeWarning),
			).String(),
		})
		if err != nil {
			logger.Warnf("Failed to list warning events for job %s/%s: %v", namespace, name, err)
		} else {
			warningEvents = eventList.Items
		}
	}
	jobStatus, failureMessage := resolveJobStatus(k8sJob, warningEvents)

	resp := &jobmodel.Job{
		Name:      k8sJob.Name,
		Namespace: k8sJob.Namespace,
		Status:    jobStatus,
		Result: &jobmodel.JobResult{
			Output: failureMessage,
		},
	}

	if k8sJob.Status.StartTime != nil {
		resp.StartTime = k8sJob.Status.StartTime.Unix()
	}
	if k8sJob.Status.CompletionTime != nil {
		resp.FinishTime = k8sJob.Status.CompletionTime.Unix()
	}

	if jobStatus == jobmodel.JobStatusSuccessful || jobStatus == jobmodel.JobStatusFailed {
		attachmentID, ok := k8sJob.Labels[bizconst.LABEL_ATTACHMENT_ID]
		if ok {
			resp.Result.AttachmentId = attachmentID
		} else {
			resp.Result.AttachmentId = fmt.Sprintf("%s.zip", k8sJob.Name)
		}
		podList, err := client.ClientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(k8sJob.Spec.Selector),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list pods")
		}
		if len(podList.Items) > 0 {
			pod := podList.Items[0]
			podLogOpts := corev1.PodLogOptions{}
			req := client.ClientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
			podLogs, err := req.Stream(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "error in opening stream")
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				return nil, errors.Wrap(err, "error in copy logs")
			}
			podOutput := buf.String()
			if resp.Result.Output != "" && podOutput != "" {
				resp.Result.Output += "\n\nPod logs:\n" + podOutput
			} else if podOutput != "" {
				resp.Result.Output = podOutput
			}

			if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].State.Terminated != nil {
				resp.Result.ExitCode = pod.Status.ContainerStatuses[0].State.Terminated.ExitCode
			}
		}
	}

	return resp, nil
}

func initialJobStatus(k8sJob *batchv1.Job) jobmodel.JobStatus {
	if k8sJob.Status.Succeeded > 0 {
		return jobmodel.JobStatusSuccessful
	}
	if k8sJob.Status.Failed > 0 {
		return jobmodel.JobStatusFailed
	}
	if k8sJob.Status.Active > 0 {
		return jobmodel.JobStatusRunning
	}
	return jobmodel.JobStatusPending
}

func resolveJobStatus(k8sJob *batchv1.Job, warningEvents []corev1.Event) (jobmodel.JobStatus, string) {
	status := initialJobStatus(k8sJob)
	for _, condition := range k8sJob.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			return jobmodel.JobStatusFailed, formatFailureMessage(condition.Reason, condition.Message)
		}
	}
	if status != jobmodel.JobStatusPending {
		return status, ""
	}
	for _, event := range warningEvents {
		if isMissingServiceAccountEvent(event) {
			return jobmodel.JobStatusFailed, formatFailureMessage(event.Reason, event.Message)
		}
	}
	return status, ""
}

func isMissingServiceAccountEvent(event corev1.Event) bool {
	if event.Type != corev1.EventTypeWarning || event.Reason != "FailedCreate" {
		return false
	}
	message := strings.ToLower(event.Message)
	return strings.Contains(message, "serviceaccount") && strings.Contains(message, "not found")
}

func formatFailureMessage(reason, message string) string {
	if reason == "" {
		return message
	}
	if message == "" {
		return reason
	}
	return fmt.Sprintf("%s: %s", reason, message)
}

func DeleteJob(ctx context.Context, namespace, name string) error {
	client := k8sclient.GetClient()
	deletePolicy := metav1.DeletePropagationBackground
	err := client.ClientSet.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}
