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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var obclusteroperationlog = logf.Log.WithName("obclusteroperation-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *OBClusterOperation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obclusteroperation,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=create;update,versions=v1alpha1,name=mobclusteroperation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBClusterOperation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBClusterOperation) Default() {
	obclusteroperationlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obclusteroperation,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=create;update,versions=v1alpha1,name=vobclusteroperation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBClusterOperation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateCreate() (admission.Warnings, error) {
	obclusteroperationlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	obclusteroperationlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateDelete() (admission.Warnings, error) {
	obclusteroperationlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
