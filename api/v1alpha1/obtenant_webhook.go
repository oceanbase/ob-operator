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

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oceanbase/ob-operator/api/constants"
)

// log is for logging in this package.
var tenantlog = logf.Log.WithName("obtenant-resource")
var tenantClt client.Client

func (r *OBTenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	tenantClt = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obtenant,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenants,verbs=create;update,versions=v1alpha1,name=mobtenant.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBTenant{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBTenant) Default() {
	cluster := &OBCluster{}
	err := tenantClt.Get(context.Background(), types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.Spec.ClusterName,
	}, cluster)
	if err != nil {
		tenantlog.Error(err, "Failed to get cluster")
	} else {
		tenantlog.Info("Get cluster", "cluster", cluster)
		r.SetOwnerReferences([]metav1.OwnerReference{{
			APIVersion: cluster.APIVersion,
			Kind:       cluster.Kind,
			Name:       cluster.GetObjectMeta().GetName(),
			UID:        cluster.GetObjectMeta().GetUID(),
		}})
	}

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
	// Ignore deleted object
	if r.GetDeletionTimestamp() != nil {
		return nil
	}
	var allErrs field.ErrorList

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

	// 3. Tenant restoring from OSS type Backup Data must have a OSSAccessSecret
	if r.Spec.Source != nil && r.Spec.Source.Restore != nil {
		res := r.Spec.Source.Restore

		if res.ArchiveSource == nil && res.BakDataSource == nil && res.SourceUri == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore"), res, "Restore must have a source option, but both archiveSource, bakDataSource and sourceUri are nil now"))
		}

		if res.ArchiveSource != nil && res.ArchiveSource.Type == constants.BackupDestTypeOSS {
			if res.ArchiveSource.OSSAccessSecret == "" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), res.ArchiveSource.OSSAccessSecret, "Tenant restoring from OSS type backup data must have a OSSAccessSecret"))
			}
			secret := &v1.Secret{}
			err := tenantClt.Get(context.Background(), types.NamespacedName{
				Namespace: r.GetNamespace(),
				Name:      res.ArchiveSource.OSSAccessSecret,
			}, secret)
			if err != nil {
				if apierrors.IsNotFound(err) {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), res.ArchiveSource.OSSAccessSecret, "Given OSSAccessSecret not found"))
				}
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), err))
			}

			if _, ok := secret.Data["accessId"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), res.ArchiveSource.OSSAccessSecret, "accessId field not found in given OSSAccessSecret"))
			}
			if _, ok := secret.Data["accessKey"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), res.ArchiveSource.OSSAccessSecret, "accessKey field not found in given OSSAccessSecret"))
			}
		}

		if res.BakDataSource != nil && res.BakDataSource.Type == constants.BackupDestTypeOSS {
			if res.BakDataSource.OSSAccessSecret == "" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "Tenant restoring from OSS type backup data must have a OSSAccessSecret"))
			}
			secret := &v1.Secret{}
			err := tenantClt.Get(context.Background(), types.NamespacedName{
				Namespace: r.GetNamespace(),
				Name:      res.BakDataSource.OSSAccessSecret,
			}, secret)
			if err != nil {
				if apierrors.IsNotFound(err) {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "Given OSSAccessSecret not found"))
				}
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), err))
			}

			if _, ok := secret.Data["accessId"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "accessId field not found in given OSSAccessSecret"))
			}
			if _, ok := secret.Data["accessKey"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "accessKey field not found in given OSSAccessSecret"))
			}
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
