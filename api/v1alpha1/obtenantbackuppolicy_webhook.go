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
	"errors"
	"regexp"

	"github.com/robfig/cron/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oceanbase/ob-operator/api/constants"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/const/status/tenantstatus"
)

// log is for logging in this package.
var backupLog = logf.Log.WithName("obtenantbackuppolicy-resource")
var bakCtl client.Client

func (r *OBTenantBackupPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	bakCtl = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obtenantbackuppolicy,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies,verbs=create;update,versions=v1alpha1,name=mobtenantbackuppolicy.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBTenantBackupPolicy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBTenantBackupPolicy) Default() {
	if r.Spec.DataBackup.Destination.Type == "" {
		r.Spec.DataBackup.Destination.Type = constants.BackupDestTypeNFS
	}
	if r.Spec.LogArchive.Destination.Type == "" {
		r.Spec.LogArchive.Destination.Type = constants.BackupDestTypeNFS
	}
	if r.Spec.LogArchive.SwitchPieceInterval == "" {
		r.Spec.LogArchive.SwitchPieceInterval = "1d"
	}
	if r.Spec.LogArchive.Binding == "" {
		r.Spec.LogArchive.Binding = constants.ArchiveBindingOptional
	}
	// only "default" is permitted
	r.Spec.DataClean.Name = "default"

	tenant := &OBTenant{}
	err := bakCtl.Get(context.Background(), types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.Spec.TenantName,
	}, tenant)
	// throw error in validator webhook
	if err != nil {
		return
	}
	if tenant.Status.Status != tenantstatus.Running {
		return
	}

	blockOwnerDeletion := true
	r.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion:         tenant.APIVersion,
		Kind:               tenant.Kind,
		Name:               tenant.GetObjectMeta().GetName(),
		UID:                tenant.GetObjectMeta().GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}})

	r.SetLabels(map[string]string{
		oceanbaseconst.LabelTenantName:   r.Spec.TenantName,
		oceanbaseconst.LabelRefOBCluster: r.Spec.ObClusterName,
		oceanbaseconst.LabelRefUID:       string(tenant.GetObjectMeta().GetUID()),
	})
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obtenantbackuppolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies,verbs=create;update,versions=v1alpha1,name=vobtenantbackuppolicy.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBTenantBackupPolicy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantBackupPolicy) ValidateCreate() (admission.Warnings, error) {
	err := r.validateBackupPolicy()
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	tenant := &OBTenant{}
	err = bakCtl.Get(ctx, types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.Spec.TenantName,
	}, tenant)
	if err != nil {
		return nil, apierrors.NewNotFound(schema.GroupResource{Group: "oceanbase.oceanbase.com", Resource: "obtenants"}, r.Spec.TenantName)
	}

	if tenant.Status.Status != tenantstatus.Running {
		return nil, errors.New("tenant is not running")
	}

	policyList := &OBTenantBackupPolicyList{}
	err = bakCtl.List(ctx, policyList, client.MatchingLabels{
		oceanbaseconst.LabelTenantName:   r.Spec.TenantName,
		oceanbaseconst.LabelRefOBCluster: r.Spec.ObClusterName,
		oceanbaseconst.LabelRefUID:       string(tenant.GetObjectMeta().GetUID()),
	})
	if err != nil {
		return nil, apierrors.NewInternalError(err)
	}
	if len(policyList.Items) > 0 {
		return nil, apierrors.NewAlreadyExists(schema.GroupResource{Group: "oceanbase.oceanbase.com", Resource: "obtenantbackuppolicies"}, policyList.Items[0].GetObjectMeta().GetName())
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantBackupPolicy) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	_ = old
	return nil, r.validateBackupPolicy()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantBackupPolicy) ValidateDelete() (admission.Warnings, error) {
	// Disabled now
	return nil, nil
}

// BackupPolicy Validation Entry
func (r *OBTenantBackupPolicy) validateBackupPolicy() error {
	if r.Spec.ObClusterName == "" {
		return errors.New("obClusterName is required")
	}
	if r.Spec.TenantName == "" {
		return errors.New("tenantName is required")
	}
	// Deprecated
	// if r.Spec.TenantSecret == "" {
	// 	return errors.New("tenantSecret is required")
	// }
	err := r.validateBackupCrontab()
	if err != nil {
		return err
	}
	err = r.validateInterval()
	if err != nil {
		return err
	}
	return nil
}

func (r *OBTenantBackupPolicy) validateInterval() error {
	var allErrs field.ErrorList
	switchPiecePattern := regexp.MustCompile(`^[1-7]d$`)
	if !switchPiecePattern.MatchString(r.Spec.LogArchive.SwitchPieceInterval) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("logArchive").Child("switchPieceInterval"), r.Spec.LogArchive.SwitchPieceInterval, "invalid switchPieceInterval"))
	}
	// RecoveryWindow will be longer than SwitchPieceInterval
	recoveryPattern := regexp.MustCompile(`^[1-9]\d*d$`)
	if !recoveryPattern.MatchString(r.Spec.DataClean.RecoveryWindow) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("dataClean").Child("recoveryWindow"), r.Spec.DataClean.RecoveryWindow, "invalid recoveryWindow"))
	}
	if r.Spec.JobKeepWindow != "" {
		jobKeepPattern := regexp.MustCompile(`^[1-9]\d*d$`)
		if !jobKeepPattern.MatchString(r.Spec.JobKeepWindow) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("jobKeepWindow"), r.Spec.JobKeepWindow, "invalid jobKeepWindow"))
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("OBTenantBackupPolicy").GroupKind(), r.Name, allErrs)
}

func (r *OBTenantBackupPolicy) validateBackupCrontab() error {
	var allErrs field.ErrorList
	err := validateScheduleFormat(r.Spec.DataBackup.FullCrontab, field.NewPath("spec").Child("dataBackup").Child("fullCrontab"))
	if err != nil {
		allErrs = append(allErrs, err)
	}
	err = validateScheduleFormat(r.Spec.DataBackup.IncrementalCrontab, field.NewPath("spec").Child("dataBackup").Child("incrementalCrontab"))
	if err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("OBTenantBackupPolicy").GroupKind(), r.Name, allErrs)
}

func validateScheduleFormat(schedule string, fldPath *field.Path) *field.Error {
	if _, err := cron.ParseStandard(schedule); err != nil {
		return field.Invalid(fldPath, schedule, err.Error())
	}
	return nil
}
