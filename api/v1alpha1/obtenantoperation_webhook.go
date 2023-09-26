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
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oceanbase/ob-operator/api/constants"
)

// log is for logging in this package.
var obtenantoperationlog = logf.Log.WithName("obtenantoperation-resource")

func (r *OBTenantOperation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obtenantoperation,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenantoperations,verbs=create;update,versions=v1alpha1,name=mobtenantoperation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBTenantOperation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBTenantOperation) Default() {
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obtenantoperation,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenantoperations,verbs=create;update,versions=v1alpha1,name=vobtenantoperation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBTenantOperation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantOperation) ValidateCreate() (admission.Warnings, error) {
	obtenantoperationlog.Info("validate create", "name", r.Name)

	return nil, r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantOperation) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	obtenantoperationlog.Info("validate update", "name", r.Name)
	_ = old
	return nil, r.validateMutation()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantOperation) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *OBTenantOperation) validateMutation() error {
	var allErrs field.ErrorList

	switch r.Spec.Type {
	case constants.TenantOpChangePwd:
		if r.Spec.ChangePwd == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("changePwd"), "change password spec is required"))
			if r.Spec.ChangePwd.SecretRef == "" || r.Spec.ChangePwd.Tenant == "" {
				allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("changePwd").Child("secretRef", "tenant"), "tenant name and secretRef are required"))
			}
		}
	case constants.TenantOpFailover:
		if r.Spec.Failover == nil || r.Spec.Failover.StandbyTenant == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("failover").Child("standbyTenant"), "name of standby tenant is activating is required"))
		}
	case constants.TenantOpSwitchover:
		if r.Spec.Switchover == nil || r.Spec.Switchover.PrimaryTenant == "" || r.Spec.Switchover.StandbyTenant == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("switchover").Child("primaryTenant", "standbyTenant"), "name of primary tenant and standby tenant are both required"))
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return allErrs.ToAggregate()
}
