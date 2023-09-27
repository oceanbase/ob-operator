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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oceanbase/ob-operator/api/constants"
)

// log is for logging in this package.
var _ = logf.Log.WithName("obtenant-resource")

func (r *OBTenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obtenant,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenants,verbs=create;update,versions=v1alpha1,name=mobtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBTenant{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBTenant) Default() {
	if r.Spec.TenantRole == "" {
		r.Spec.TenantRole = constants.TenantRolePrimary
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obtenant,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenants,verbs=create;update,versions=v1alpha1,name=vobtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBTenant{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenant) ValidateCreate() (admission.Warnings, error) {
	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenant) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	_ = old
	// TODO(user): fill in your validation logic upon object update.
	return nil, r.validateMutation()
}

func (r *OBTenant) validateMutation() error {
	var allErrs field.ErrorList

	if r.Spec.Credentials.Root == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("credentials").Child("root"), r.Spec.Credentials.Root, "Root password user secretref must be set"))
	}
	if r.Spec.Credentials.StandbyRO == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("credentials").Child("standbyRo"), r.Spec.Credentials.StandbyRO, "Standby read-only user password secretref must be set"))
	}

	// 1. Standby tenant must have a source
	if r.Spec.TenantRole == constants.TenantRoleStandby {
		if r.Spec.Source == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source"), r.Spec.Source, "Standby tenant must have non-nil source field"))
		} else if r.Spec.Source.Restore == nil && r.Spec.Source.Tenant == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("tenantRole"), r.Spec.TenantRole, "Standby must have a source option, but both restore and tenantRef are nil now"))
		}
	}

	// 2. Restore until with some limit must have a limit key
	if r.Spec.Source != nil && r.Spec.Source.Restore != nil {
		untilSpec := r.Spec.Source.Restore.Until
		if !untilSpec.Unlimited && untilSpec.Scn == nil && untilSpec.Timestamp == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("until"), untilSpec, "Restore until must have a limit key, scn and timestamp are both nil now"))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("OBTenant").GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenant) ValidateDelete() (admission.Warnings, error) {
	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
