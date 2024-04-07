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
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
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
		clusterMeta := cluster.GetObjectMeta()
		r.SetOwnerReferences([]metav1.OwnerReference{{
			APIVersion: cluster.APIVersion,
			Kind:       cluster.Kind,
			Name:       clusterMeta.GetName(),
			UID:        clusterMeta.GetUID(),
		}})
		labels := r.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[oceanbaseconst.LabelRefOBCluster] = clusterMeta.GetName()
		r.SetLabels(labels)
	}

	if r.Spec.TenantRole == "" {
		r.Spec.TenantRole = constants.TenantRolePrimary
	} else {
		r.Spec.TenantRole = apitypes.TenantRole(strings.ToUpper(string(r.Spec.TenantRole)))
	}

	if r.Spec.Credentials.StandbyRO == "" {
		r.Spec.Credentials.StandbyRO = "standby-ro-" + rand.String(8)
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
	if r.Status.Status == tenantstatus.Running {
		switch {
		case r.Spec.ClusterName != old.(*OBTenant).Spec.ClusterName:
			return nil, apierrors.NewBadRequest("Cannot change clusterName when tenant is running")
		case r.Spec.TenantName != old.(*OBTenant).Spec.TenantName:
			return nil, apierrors.NewBadRequest("Cannot change tenantName when tenant is running")
		}
	}
	return nil, r.validateMutation()
}

func (r *OBTenant) validateMutation() error {
	// Ignore deleted object
	if r.GetDeletionTimestamp() != nil {
		return nil
	}
	var allErrs field.ErrorList

	// Check the unit number
	if r.Spec.UnitNumber <= 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("unitNum"), r.Spec.UnitNumber, "unitNum must be greater than 0"))
	}

	// Check the legality of tenantName
	tenantNamePattern := regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]{0,127}$")
	if !tenantNamePattern.MatchString(r.Spec.TenantName) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("tenantName"), r.Spec.TenantName, "Invalid tenantName, which should start with character or underscore and contain character, digit and underscore only"))
	}

	// TenantRole must be one of PRIMARY and STANDBY
	if r.Spec.TenantRole != constants.TenantRolePrimary && r.Spec.TenantRole != constants.TenantRoleStandby {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("tenantRole"), r.Spec.TenantRole, "TenantRole must be primary or standby"))
	}

	// OBCluster must exist
	cluster := &OBCluster{}
	err := tenantClt.Get(context.Background(), types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.Spec.ClusterName,
	}, cluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("clusterName"), r.Spec.ClusterName, "Given cluster not found"))
		} else {
			allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("clusterName"), err))
		}
	} else {
		// Check whether zones in tenant.spec.pools exist or not
		for i, pool := range r.Spec.Pools {
			exist := false
			for _, zone := range cluster.Spec.Topology {
				if pool.Zone == zone.Zone {
					exist = true
					break
				}
			}
			if !exist {
				msg := fmt.Sprintf("Zone %s does not exist in cluster %s", pool.Zone, cluster.Spec.ClusterName)
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("pools").Child(fmt.Sprintf("%d", i)), pool.Zone, msg))
			}
		}
	}

	// Given credentials must exist
	if r.Spec.Credentials.Root != "" {
		secret := &v1.Secret{}
		err = tenantClt.Get(context.Background(), types.NamespacedName{
			Namespace: r.GetNamespace(),
			Name:      r.Spec.Credentials.Root,
		}, secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("credentials").Child("root"), r.Spec.Credentials.Root, "Given root credential not found"))
			} else {
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("credentials").Child("root"), err))
			}
		} else {
			if _, ok := secret.Data["password"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("credentials").Child("standbyRo"), r.Spec.Credentials.StandbyRO, "password field not found in given standbyRo credential"))
			}
		}
	}

	if r.Spec.Credentials.StandbyRO != "" {
		secret := &v1.Secret{}
		err = tenantClt.Get(context.Background(), types.NamespacedName{
			Namespace: r.GetNamespace(),
			Name:      r.Spec.Credentials.StandbyRO,
		}, secret)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("credentials").Child("standbyRo"), err))
			}
		} else {
			if _, ok := secret.Data["password"]; !ok {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("credentials").Child("standbyRo"), r.Spec.Credentials.StandbyRO, "password field not found in given standbyRo credential"))
			}
		}
	}

	// 1. Standby tenant must have a source; source.tenant must be valid
	if r.Spec.TenantRole == constants.TenantRoleStandby {
		if r.Spec.Source == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source"), r.Spec.Source, "Standby tenant must have non-nil source field"))
		} else if r.Spec.Source.Restore == nil && r.Spec.Source.Tenant == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("tenantRole"), r.Spec.TenantRole, "Standby must have a source option, but both restore and tenantRef are nil now"))
		} else if r.Spec.Source.Tenant != nil {
			tenant := &OBTenant{}
			ns := r.GetNamespace()
			tenantCR := *r.Spec.Source.Tenant
			splits := strings.Split(*r.Spec.Source.Tenant, "/")
			switch len(splits) {
			case 0, 1:
			case 2:
				if splits[0] == "" {
					return field.Invalid(field.NewPath("spec").Child("source").Child("tenant"), tenantCR, "Given tenant namespace is empty")
				}
				if splits[1] == "" {
					return field.Invalid(field.NewPath("spec").Child("source").Child("tenant"), tenantCR, "Given tenant name is empty")
				}
				ns = splits[0]
				tenantCR = splits[1]
			default:
				return field.Invalid(field.NewPath("spec").Child("source").Child("tenant"), tenantCR, "Given tenant name is invalid, it should be namespace/name or name format")
			}
			err = tenantClt.Get(context.TODO(), types.NamespacedName{
				Namespace: ns,
				Name:      tenantCR,
			}, tenant)
			if err != nil {
				if apierrors.IsNotFound(err) {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("tenant"), tenantCR, "Given tenant not found in namespace "+ns))
				} else {
					allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("source").Child("tenant"), err))
				}
			}
			cluster := &OBCluster{}
			err = tenantClt.Get(context.Background(), types.NamespacedName{
				Namespace: ns,
				Name:      tenant.Spec.ClusterName,
			}, cluster)
			if err != nil {
				if apierrors.IsNotFound(err) {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("tenant"), tenantCR, "Given tenant not found in namespace "+ns))
				} else {
					allErrs = append(allErrs, field.InternalError(field.NewPath("spec").Child("source").Child("tenant"), err))
				}
			}
			clusterAnnotations := cluster.GetAnnotations()
			if clusterAnnotations != nil {
				if mode, exist := clusterAnnotations[oceanbaseconst.AnnotationsMode]; exist && mode == oceanbaseconst.ModeStandalone {
					allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("tenant"), r.Spec.Source.Tenant, "Given tenant is in a standalone cluster, which can not be a restore source"))
				}
			}
		}
	}

	if len(allErrs) > 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind("OBTenant").GroupKind(), r.Name, allErrs)
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
			} else {
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
				} else {
					if _, ok := secret.Data["accessId"]; !ok {
						allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), res.ArchiveSource.OSSAccessSecret, "accessId field not found in given OSSAccessSecret"))
					}
					if _, ok := secret.Data["accessKey"]; !ok {
						allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("archiveSource").Child("ossAccessSecret"), res.ArchiveSource.OSSAccessSecret, "accessKey field not found in given OSSAccessSecret"))
					}
				}
			}
		}

		if res.BakDataSource != nil && res.BakDataSource.Type == constants.BackupDestTypeOSS {
			if res.BakDataSource.OSSAccessSecret == "" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "Tenant restoring from OSS type backup data must have a OSSAccessSecret"))
			} else {
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
				} else {
					if _, ok := secret.Data["accessId"]; !ok {
						allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "accessId field not found in given OSSAccessSecret"))
					}
					if _, ok := secret.Data["accessKey"]; !ok {
						allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("source").Child("restore").Child("bakDataSource").Child("ossAccessSecret"), res.BakDataSource.OSSAccessSecret, "accessKey field not found in given OSSAccessSecret"))
					}
				}
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
