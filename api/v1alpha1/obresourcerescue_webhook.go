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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var obresourcerescuelog = logf.Log.WithName("obresourcerescue-resource")

var rescueTypeMapping = map[string]struct{}{
	"delete":          {},
	"reset":           {},
	"retry":           {},
	"skip":            {},
	"ignore-deletion": {},
}

func (r *OBResourceRescue) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obresourcerescue,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obresourcerescues,verbs=create;update,versions=v1alpha1,name=mobresourcerescue.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBResourceRescue{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBResourceRescue) Default() {
	r.Spec.Type = strings.ToLower(r.Spec.Type)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obresourcerescue,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obresourcerescues,verbs=create;update,versions=v1alpha1,name=vobresourcerescue.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBResourceRescue{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBResourceRescue) ValidateCreate() (admission.Warnings, error) {
	return r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBResourceRescue) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldRes := old.(*OBResourceRescue)
	if r.Status.Status == "Successful" || r.Status.Status == "" {
		if r.Spec.Type != oldRes.Spec.Type {
			return nil, field.Invalid(field.NewPath("spec", "type"), r.Spec.Type, "type cannot be changed")
		}
		if r.Spec.TargetKind != oldRes.Spec.TargetKind {
			return nil, field.Invalid(field.NewPath("spec", "targetKind"), r.Spec.TargetKind, "targetKind cannot be changed")
		}
		if r.Spec.TargetResName != oldRes.Spec.TargetResName {
			return nil, field.Invalid(field.NewPath("spec", "targetResName"), r.Spec.TargetResName, "targetResName cannot be changed")
		}
		if r.Spec.TargetGV != oldRes.Spec.TargetGV {
			return nil, field.Invalid(field.NewPath("spec", "targetGV"), r.Spec.TargetGV, "targetGV cannot be changed")
		}
		if r.Spec.Namespace != oldRes.Spec.Namespace {
			return nil, field.Invalid(field.NewPath("spec", "namespace"), r.Spec.Namespace, "namespace cannot be changed")
		}
		if r.Spec.TargetStatus != oldRes.Spec.TargetStatus {
			return nil, field.Invalid(field.NewPath("spec", "targetStatus"), r.Spec.TargetStatus, "targetStatus cannot be changed")
		}
	}
	return r.validateMutation()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBResourceRescue) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *OBResourceRescue) validateMutation() (admission.Warnings, error) {
	var errList field.ErrorList
	var warnings []string

	if r.Spec.TargetKind == "" {
		errList = append(errList, field.Required(field.NewPath("spec", "targetKind"), "targetKind is required"))
	}
	if r.Spec.TargetResName == "" {
		errList = append(errList, field.Required(field.NewPath("spec", "targetResName"), "targetResName is required"))
	}
	if r.Spec.Type == "" {
		errList = append(errList, field.Required(field.NewPath("spec", "type"), "type is required"))
	} else if _, exist := rescueTypeMapping[r.Spec.Type]; !exist {
		errList = append(errList, field.Invalid(field.NewPath("spec", "type"), r.Spec.Type, "unsupported rescue type"))
	} else if r.Spec.Type == "reset" && r.Spec.TargetStatus == "" {
		errList = append(errList, field.Required(field.NewPath("spec", "targetStatus"), "targetStatus is required when type is reset"))
	}

	return warnings, errList.ToAggregate()
}
