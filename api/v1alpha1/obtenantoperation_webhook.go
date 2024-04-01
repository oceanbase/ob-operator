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
	"strings"

	v1 "k8s.io/api/core/v1"
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
	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

// log is for logging in this package.
var obtenantoperationlog = logf.Log.WithName("obtenantoperation-resource")
var clt client.Client

func (r *OBTenantOperation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	clt = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obtenantoperation,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenantoperations,verbs=create;update,versions=v1alpha1,name=mobtenantoperation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBTenantOperation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBTenantOperation) Default() {
	r.Spec.Type = apitypes.TenantOperationType(strings.ToUpper(string(r.Spec.Type)))

	tenant := &OBTenant{}
	var targetTenantName string
	var secondaryTenantName string
	if r.Spec.Type == constants.TenantOpChangePwd && r.Spec.ChangePwd != nil {
		targetTenantName = r.Spec.ChangePwd.Tenant
	} else if r.Spec.Type == constants.TenantOpFailover && r.Spec.Failover != nil {
		targetTenantName = r.Spec.Failover.StandbyTenant
	} else if r.Spec.Type == constants.TenantOpSwitchover && r.Spec.Switchover != nil {
		targetTenantName = r.Spec.Switchover.PrimaryTenant
		secondaryTenantName = r.Spec.Switchover.StandbyTenant
	} else if (r.Spec.Type == constants.TenantOpUpgrade || r.Spec.Type == constants.TenantOpReplayLog) && r.Spec.TargetTenant != nil {
		targetTenantName = *r.Spec.TargetTenant
	}
	references := r.GetOwnerReferences()
	labels := r.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	if targetTenantName != "" {
		err := clt.Get(context.Background(), types.NamespacedName{
			Namespace: r.GetNamespace(),
			Name:      targetTenantName,
		}, tenant)
		if err != nil {
			// obtenantoperationlog.Error(err, "get tenant")
			return
		}
		firstMeta := tenant.GetObjectMeta()
		references = append(references, metav1.OwnerReference{
			APIVersion: tenant.APIVersion,
			Kind:       tenant.Kind,
			Name:       firstMeta.GetName(),
			UID:        firstMeta.GetUID(),
		})
		labels[oceanbaseconst.LabelTenantName] = firstMeta.GetName()
	}

	if secondaryTenantName != "" {
		secondaryTenant := &OBTenant{}
		err := clt.Get(context.Background(), types.NamespacedName{
			Namespace: r.GetNamespace(),
			Name:      secondaryTenantName,
		}, secondaryTenant)
		if err != nil {
			// obtenantoperationlog.Error(err, "get tenant")
			return
		}
		secondMeta := secondaryTenant.GetObjectMeta()
		references = append(references, metav1.OwnerReference{
			APIVersion: secondaryTenant.APIVersion,
			Kind:       secondaryTenant.Kind,
			Name:       secondMeta.GetName(),
			UID:        secondMeta.GetUID(),
		})
		labels[oceanbaseconst.LabelSecondaryTenant] = secondMeta.GetName()
	}

	r.SetOwnerReferences(references)
	r.SetLabels(labels)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obtenantoperation,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obtenantoperations,verbs=create;update,versions=v1alpha1,name=vobtenantoperation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBTenantOperation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantOperation) ValidateCreate() (admission.Warnings, error) {
	return nil, r.validateMutation()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBTenantOperation) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	_ = old
	warnings := []string{"Updating operation resource can not trigger any action, please create a new one if you want to do that"}
	return warnings, r.validateMutation()
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
		} else if r.Spec.ChangePwd.SecretRef == "" || r.Spec.ChangePwd.Tenant == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("changePwd").Child("secretRef", "tenant"), "tenant name and secretRef are required"))
		} else if _, err := r.checkTenantCRExistence(r.Spec.ChangePwd.Tenant); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("changePwd").Child("tenant"), r.Spec.ChangePwd.Tenant, "Failed to get tenant of given name"))
		} else {
			sec, err := r.checkSecretExistence(r.GetNamespace(), r.Spec.ChangePwd.SecretRef)
			if err != nil {
				if apierrors.IsNotFound(err) {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("changePwd").Child("secretRef"), r.Spec.ChangePwd.SecretRef, "Given secret not found"))
				} else {
					allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("changePwd").Child("secretRef"), err))
				}
			} else if _, ok := sec.Data["password"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("changePwd").Child("secretRef"), r.Spec.ChangePwd.SecretRef, "'password' field not found in data of secret"))
			}
		}
	case constants.TenantOpFailover:
		if r.Spec.Failover == nil || r.Spec.Failover.StandbyTenant == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("failover").Child("standbyTenant"), "name of standby tenant is activating is required"))
		} else {
			tenant, err := r.checkTenantCRExistence(r.Spec.Failover.StandbyTenant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("failover").Child("standbyTenant"), r.Spec.Failover.StandbyTenant, "Failed to get standby tenant of given name"))
			} else if tenant.Status.TenantRole != constants.TenantRoleStandby {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("failover").Child("standbyTenant"), r.Spec.Failover.StandbyTenant, fmt.Sprintf("Tenant %s is not a standby tenant", r.Spec.Failover.StandbyTenant)))
			}
		}
	case constants.TenantOpSwitchover:
		if r.Spec.Switchover == nil || r.Spec.Switchover.PrimaryTenant == "" || r.Spec.Switchover.StandbyTenant == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("switchover").Child("primaryTenant", "standbyTenant"), "name of primary tenant and standby tenant are both required"))
		} else {
			primary, err := r.checkTenantCRExistence(r.Spec.Switchover.PrimaryTenant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("switchover").Child("primaryTenant"), r.Spec.Switchover.PrimaryTenant, "Failed to get primary tenant of given name"))
			} else if primary.Status.TenantRole != constants.TenantRolePrimary {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("switchover").Child("primaryTenant"), r.Spec.Switchover.PrimaryTenant, fmt.Sprintf("Tenant %s is not a primary tenant", r.Spec.Switchover.PrimaryTenant)))
			} else if primary.Status.Status != "running" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("switchover").Child("primaryTenant"), r.Spec.Switchover.PrimaryTenant, "The primary tenant is not in running status"))
			}
			standby, err := r.checkTenantCRExistence(r.Spec.Switchover.StandbyTenant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("switchover").Child("standbyTenant"), r.Spec.Switchover.StandbyTenant, "Failed to get standby tenant of given name"))
			} else if standby.Status.TenantRole != constants.TenantRoleStandby {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("switchover").Child("standbyTenant"), r.Spec.Switchover.StandbyTenant, fmt.Sprintf("Tenant %s is not a standby tenant", r.Spec.Switchover.StandbyTenant)))
			} else if standby.Status.Status != "running" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("switchover").Child("standbyTenant"), r.Spec.Switchover.StandbyTenant, "The standby tenant is not in running status"))
			}
		}
	case constants.TenantOpUpgrade:
		if r.Spec.TargetTenant == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("targetTenant"), "name of targetTenant is required"))
		} else {
			tenant, err := r.checkTenantCRExistence(*r.Spec.TargetTenant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("targetTenant"), r.Spec.TargetTenant, "Failed to get target tenant of given name"))
			} else if tenant.Status.TenantRole != constants.TenantRolePrimary {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("targetTenant"), r.Spec.TargetTenant, "Standby tenant cannot be upgraded, please activate it first"))
			}
		}
	case constants.TenantOpReplayLog:
		untilSpec := r.Spec.ReplayUntil
		if untilSpec == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("replayUntil"), "replayUntil is required"))
		} else if !untilSpec.Unlimited && untilSpec.Scn == nil && untilSpec.Timestamp == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("replayUntil"), untilSpec, "Limited replayUntil must have a limit key, scn and timestamp are both nil now"))
		}
		if r.Spec.TargetTenant == nil {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("targetTenant"), "name of targetTenant is required"))
		} else {
			tenant, err := r.checkTenantCRExistence(*r.Spec.TargetTenant)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("targetTenant"), r.Spec.TargetTenant, "Failed to get target tenant of given name"))
			} else if tenant.Status.TenantRole != constants.TenantRoleStandby {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("targetTenant"), r.Spec.TargetTenant, "The target tenant is not a standby"))
			}
		}
	default:
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("type"), string(r.Spec.Type)+" type of operation is not supported"))
	}
	if len(allErrs) == 0 {
		return nil
	}
	return allErrs.ToAggregate()
}

func (r *OBTenantOperation) checkTenantCRExistence(tenantCRName string) (*OBTenant, error) {
	tenant := &OBTenant{}
	err := clt.Get(context.TODO(), types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      tenantCRName,
	}, tenant)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, apierrors.NewNotFound(schema.GroupResource{
				Group:    "oceanbase.oceanbase.com",
				Resource: "OBTenant",
			}, tenantCRName)
		}
		return nil, apierrors.NewInternalError(err)
	}
	return tenant, nil
}

func (r *OBTenantOperation) checkSecretExistence(ns, secretName string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	err := tenantClt.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      secretName,
	}, secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}
