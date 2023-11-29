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
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var obclusterlog = logf.Log.WithName("obcluster-resource")

func (r *OBCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusters,verbs=create;update,versions=v1alpha1,name=mobcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBCluster) Default() {
	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusters,verbs=create;update,versions=v1alpha1,name=vobcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateCreate() (admission.Warnings, error) {
	obclusterlog.Info("validate create", "name", r.Name)

	return nil, r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	_ = old
	obclusterlog.Info("validate update", "name", r.Name)

	return nil, r.validateMutation()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBCluster) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *OBCluster) validateMutation() error {
	// Ignore deleting objects
	if r.GetDeletionTimestamp() != nil {
		return nil
	}

	var allErrs field.ErrorList

	// 0. Validate userSecrets
	if r.Spec.UserSecrets != nil {
		if err := r.checkSecretExistence(r.Namespace, r.Spec.UserSecrets.Root, "root"); err != nil {
			allErrs = append(allErrs, err)
		}
		if err := r.checkSecretExistence(r.Namespace, r.Spec.UserSecrets.ProxyRO, "proxyro"); err != nil {
			allErrs = append(allErrs, err)
		}
		if err := r.checkSecretExistence(r.Namespace, r.Spec.UserSecrets.Operator, "operator"); err != nil {
			allErrs = append(allErrs, err)
		}
		if err := r.checkSecretExistence(r.Namespace, r.Spec.UserSecrets.Monitor, "monitor"); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	// 1. Validate Topology
	if r.Spec.Topology == nil || len(r.Spec.Topology) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("topology"), r.Spec.Topology, "empty topology is not permitted"))
	}

	// 2. Validate storageClasses
	storageClassMapping := make(map[string]bool)
	storageClassMapping[r.Spec.OBServerTemplate.Storage.DataStorage.StorageClass] = true
	storageClassMapping[r.Spec.OBServerTemplate.Storage.LogStorage.StorageClass] = true
	storageClassMapping[r.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass] = true

	for key := range storageClassMapping {
		err := clt.Get(context.TODO(), types.NamespacedName{
			Name: key,
		}, &storagev1.StorageClass{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("storage").Child("storageClass"), key, fmt.Sprintf("storageClass %s not found", key)))
			} else {
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("observer").Child("storage").Child("storageClass"), err))
			}
		}
	}

	// 3. Validate memory size
	clusterMinMemory := resource.MustParse("8Gi")
	if r.Spec.MonitorTemplate.Resource.Memory.AsApproximateFloat64() < clusterMinMemory.AsApproximateFloat64() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("observer").Child("resource").Child("memory"), r.Spec.MonitorTemplate.Resource.Memory.String(), "The minimum memory size of OBCluster is "+clusterMinMemory.String()))
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("OBCluster").GroupKind(), r.Name, allErrs)
}

func (r *OBCluster) checkSecretExistence(ns, secretName, fieldName string) *field.Error {
	if secretName == "" {
		return field.Invalid(field.NewPath("spec").Child("userSecrets").Child(fieldName), secretName, fmt.Sprintf("Empty credential %s is not permitted", fieldName))
	}
	secret := &v1.Secret{}
	err := tenantClt.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      secretName,
	}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return field.Invalid(field.NewPath("spec").Child("userSecrets").Child(fieldName), secretName, fmt.Sprintf("Given %s credential %s not found", fieldName, secretName))
		}
		return field.InternalError(field.NewPath("spec").Child("userSecrets").Child(fieldName), err)
	}
	return nil
}
