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
	"context"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func GetJob(ctx context.Context, namespace, name string) (*job.Job, error) {
	client := k8sclient.GetClient()
	k8sJob, err := client.ClientSet.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &job.Job{
		Name:       k8sJob.Name,
		Namespace:  k8sJob.Namespace,
		Status:     job.JobStatus(k8sJob.Status.Conditions[0].Type),
		StartTime:  k8sJob.Status.StartTime.Unix(),
		FinishTime: k8sJob.Status.CompletionTime.Unix(),
	}, nil
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
