/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package resource

import (
	"context"

	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobResource struct {
	*Resource
}

func NewJobResource(resource *Resource) ResourceOperator {
	return &JobResource{resource}
}

func (r *JobResource) Create(ctx context.Context, obj interface{}) error {
	Job := obj.(batchv1.Job)
	// kube.LogForAppActionStatus(Job.Kind, Job.Name, "create", Job)
	err := r.Client.Create(ctx, &Job)
	if err != nil {
		r.Recorder.Eventf(&Job, corev1.EventTypeWarning, FailedToCreateJob, "create Job"+Job.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(Job.Kind, Job.Name, "create", "succeed")
	r.Recorder.Event(&Job, corev1.EventTypeNormal, CreatedJob, "create Job"+Job.Name)
	return nil
}

func (r *JobResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	Job := &batchv1.Job{}
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), Job)
	if err != nil {
		klog.Errorln(err)
	}
	return *Job, err
}

func (r *JobResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	JobList := &batchv1.JobList{}
	err := r.Client.List(ctx, JobList, client.InNamespace(namespace), listOption)
	if err != nil {
		// can definitely get a value, so errors are not returned
		klog.Errorln(err)
	}
	return *JobList
}

func (r *JobResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *JobResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	var res error
	return res
}

func (r *JobResource) Delete(ctx context.Context, obj interface{}) error {
	Job := obj.(batchv1.Job)
	// kube.LogForAppActionStatus(Job.Kind, Job.Name, "delete", Job)
	err := r.Client.Delete(ctx, &Job)
	if err != nil {
		r.Recorder.Eventf(&Job, corev1.EventTypeWarning, FailedToKillJob, "delete Job"+Job.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(Job.Kind, Job.Name, "delete", "succeed")
	r.Recorder.Event(&Job, corev1.EventTypeNormal, DeletedJob, "delete Job"+Job.Name)
	return nil
}

func (r *JobResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}

func (r *JobResource) PatchStatus(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
