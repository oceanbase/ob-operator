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
	"io"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJob(ctx context.Context, namespace, name string) (*job.Job, error) {
	client := k8sclient.GetClient()
	k8sJob, err := client.ClientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jobStatus := job.JobStatusPending
	if k8sJob.Status.Succeeded > 0 {
		jobStatus = job.JobStatusSuccessful
	} else if k8sJob.Status.Failed > 0 {
		jobStatus = job.JobStatusFailed
	} else if k8sJob.Status.Active > 0 {
		jobStatus = job.JobStatusRunning
	}

	resp := &job.Job{
		Name:      k8sJob.Name,
		Namespace: k8sJob.Namespace,
		Status:    jobStatus,
		Result:    &job.JobResult{},
	}

	if k8sJob.Status.StartTime != nil {
		resp.StartTime = k8sJob.Status.StartTime.Unix()
	}
	if k8sJob.Status.CompletionTime != nil {
		resp.FinishTime = k8sJob.Status.CompletionTime.Unix()
	}

	if jobStatus == job.JobStatusSuccessful || jobStatus == job.JobStatusFailed {
		attachmentID, ok := k8sJob.Labels[alarm.DIAGNOSE_LABEL_ATTACHMENT_ID]
		if ok {
			resp.Result.AttachmentId = attachmentID
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
			resp.Result.Output = buf.String()

			if pod.Status.ContainerStatuses[0].State.Terminated != nil {
				resp.Result.ExitCode = pod.Status.ContainerStatuses[0].State.Terminated.ExitCode
			}
		}
	}

	return resp, nil
}

func DeleteJob(ctx context.Context, namespace, name string) error {
	client := k8sclient.GetClient()
	err := client.ClientSet.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}
