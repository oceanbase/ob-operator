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
	"strings"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/oceanbase/ob-operator/api/constants"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-obclusteroperation,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=create;update,versions=v1alpha1,name=mobclusteroperation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBClusterOperation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OBClusterOperation) Default() {
	ctx := context.Background()
	obcluster := OBCluster{}
	err := clt.Get(ctx, types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Spec.OBCluster,
	}, &obcluster)
	if err != nil {
		obclusteroperationlog.Info("obcluster not found", "name", r.Spec.OBCluster)
		return
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obclusteroperation,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=create;update,versions=v1alpha1,name=vobclusteroperation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBClusterOperation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateCreate() (admission.Warnings, error) {
	obclusteroperationlog.Info("validate create", "name", r.Name)
	if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeAddZones)) && r.Spec.AddZones == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("addZones"), r.Spec.AddZones, "addZones must be set for cluster operation of type addZones")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeDeleteZones)) && r.Spec.DeleteZones == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("deleteZones"), r.Spec.DeleteZones, "deleteZones must be set for cluster operation of type deleteZones")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeAdjustReplicas)) && r.Spec.AdjustReplicas == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("adjustReplicas"), r.Spec.AdjustReplicas, "adjustReplicas must be set for cluster operation of type adjustReplicas")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeUpgrade)) && r.Spec.Upgrade == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("upgrade"), r.Spec.Upgrade, "upgrade must be set for cluster operation of type upgrade")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeRestartOBServers)) && r.Spec.RestartOBServers == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("restartOBServers"), r.Spec.RestartOBServers, "restartOBServers must be set for cluster operation of type restartOBServers")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeModifyStorageClass)) && r.Spec.ModifyStorageClass == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass"), r.Spec.ModifyStorageClass, "modifyStorageClass must be set for cluster operation of type modifyStorageClass")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeExpandStorageSize)) && r.Spec.ExpandStorageSize == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageSize"), r.Spec.ExpandStorageSize, "modifyStorageSize must be set for cluster operation of type modifyStorageSize")
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeSetParameters)) && r.Spec.SetParameters == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("setParameters"), r.Spec.SetParameters, "setParameters must be set for cluster operation of type setParameters")
	}

	ctx := context.Background()
	obcluster := OBCluster{}
	err := clt.Get(ctx, types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Spec.OBCluster,
	}, &obcluster)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, field.Invalid(field.NewPath("spec").Child("obcluster"), r.Spec.OBCluster, "obcluster not found")
		}
		return nil, kubeerrors.NewInternalError(err)
	}
	if !r.Spec.Force && obcluster.Status.Status != "running" {
		return nil, field.Invalid(field.NewPath("spec").Child("obcluster"), r.Spec.OBCluster, "obcluster is currently operating, please use force to override")
	}
	if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeExpandStorageSize)) && r.Spec.ExpandStorageSize != nil {
		if r.Spec.ExpandStorageSize.DataStorage.Cmp(obcluster.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 {
			return nil, field.Invalid(field.NewPath("spec").Child("expandStorageSize").Child("dataStorage"), r.Spec.ExpandStorageSize, "storage size can not be less than current size")
		}
		if r.Spec.ExpandStorageSize.LogStorage.Cmp(obcluster.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 {
			return nil, field.Invalid(field.NewPath("spec").Child("expandStorageSize").Child("logStorage"), r.Spec.ExpandStorageSize, "storage size can not be less than current size")
		}
		if r.Spec.ExpandStorageSize.RedoLogStorage.Cmp(obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0 {
			return nil, field.Invalid(field.NewPath("spec").Child("expandStorageSize").Child("redoLogStorage"), r.Spec.ExpandStorageSize, "storage size can not be less than current size")
		}
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeModifyStorageClass)) && r.Spec.ModifyStorageClass != nil {
		if r.Spec.ModifyStorageClass.DataStorage != "" &&
			r.Spec.ModifyStorageClass.DataStorage != obcluster.Spec.OBServerTemplate.Storage.DataStorage.StorageClass &&
			validateStorageClassAllowExpansion(r.Spec.ModifyStorageClass.DataStorage) != nil {
			return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass").Child("dataStorage"), r.Spec.ModifyStorageClass, "storage class does not support expansion")
		}
		if r.Spec.ModifyStorageClass.LogStorage != "" &&
			r.Spec.ModifyStorageClass.LogStorage != obcluster.Spec.OBServerTemplate.Storage.LogStorage.StorageClass &&
			validateStorageClassAllowExpansion(r.Spec.ModifyStorageClass.LogStorage) != nil {
			return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass").Child("logStorage"), r.Spec.ModifyStorageClass, "storage class does not support expansion")
		}
		if r.Spec.ModifyStorageClass.RedoLogStorage != "" &&
			r.Spec.ModifyStorageClass.RedoLogStorage != obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass &&
			validateStorageClassAllowExpansion(r.Spec.ModifyStorageClass.RedoLogStorage) != nil {
			return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass").Child("redoLogStorage"), r.Spec.ModifyStorageClass, "storage class does not support expansion")
		}
	} else if strings.EqualFold(string(r.Spec.Type), string(constants.ClusterOpTypeRestartOBServers)) &&
		obcluster.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] != "true" {
		return nil, field.Invalid(field.NewPath("spec").Child("obcluster"), r.Spec.OBCluster, "obcluster does not support static ip, can not restart observers")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	warnings := []string{"Update to OBClusterOperation won't take effect."}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateDelete() (admission.Warnings, error) {
	obclusteroperationlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
