/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sv1alpha1

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// log is for logging in this package.
var k8sclusterlog = logf.Log.WithName("k8scluster-resource")

func (r *K8sCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-k8s-oceanbase-com-v1alpha1-k8scluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=k8s.oceanbase.com,resources=k8sclusters,verbs=create;update,versions=v1alpha1,name=mk8scluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &K8sCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *K8sCluster) Default() {
	k8sclusterlog.Info("default", "name", r.Name)
	r.Spec.KubeConfig = r.EncodeKubeConfig()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-k8s-oceanbase-com-v1alpha1-k8scluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=k8s.oceanbase.com,resources=k8sclusters,verbs=create;update,versions=v1alpha1,name=vk8scluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &K8sCluster{}

func (r *K8sCluster) validateMutation() (admission.Warnings, error) {
	kubeConfig, err := r.DecodeKubeConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode kubeconfig")
	}
	config, err := k8sclient.GetConfigFromBytes(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config from kubeconfig field of %s", r.Name)
	}
	client, err := k8sclient.GetCtrlRuntimeClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create k8s client")
	}
	podList := corev1.PodList{}
	err = client.List(context.TODO(), &podList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods with given kubeconfig")
	}
	return nil, nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *K8sCluster) ValidateCreate() (admission.Warnings, error) {
	k8sclusterlog.Info("validate create", "name", r.Name)
	return r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *K8sCluster) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	k8sclusterlog.Info("validate update", "name", r.Name)
	return r.validateMutation()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *K8sCluster) ValidateDelete() (admission.Warnings, error) {
	k8sclusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
