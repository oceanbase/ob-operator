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

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
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
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[oceanbaseconst.LabelRefOBCluster] = obcluster.Name
	r.Spec.Type = apitypes.ClusterOperationType(strings.ToUpper(string(r.Spec.Type[0])) + string(r.Spec.Type[1:]))
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-obclusteroperation,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=create;update,versions=v1alpha1,name=vobclusteroperation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBClusterOperation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OBClusterOperation) ValidateCreate() (admission.Warnings, error) {
	obclusteroperationlog.Info("validate create", "name", r.Name)
	if r.Spec.Type == constants.ClusterOpTypeAddZones && r.Spec.AddZones == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("addZones"), r.Spec.AddZones, "addZones must be set for cluster operation of type addZones")
	} else if r.Spec.Type == constants.ClusterOpTypeDeleteZones && r.Spec.DeleteZones == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("deleteZones"), r.Spec.DeleteZones, "deleteZones must be set for cluster operation of type deleteZones")
	} else if r.Spec.Type == constants.ClusterOpTypeAdjustReplicas && r.Spec.AdjustReplicas == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("adjustReplicas"), r.Spec.AdjustReplicas, "adjustReplicas must be set for cluster operation of type adjustReplicas")
	} else if r.Spec.Type == constants.ClusterOpTypeUpgrade && r.Spec.Upgrade == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("upgrade"), r.Spec.Upgrade, "upgrade must be set for cluster operation of type upgrade")
	} else if r.Spec.Type == constants.ClusterOpTypeRestartOBServers && r.Spec.RestartOBServers == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("restartOBServers"), r.Spec.RestartOBServers, "restartOBServers must be set for cluster operation of type restartOBServers")
	} else if r.Spec.Type == constants.ClusterOpTypeModifyOBServers && r.Spec.ModifyOBServers == nil {
		return nil, field.Invalid(field.NewPath("spec").Child("modifyOBServers"), r.Spec.ModifyOBServers, "modifyOBServers must be set for cluster operation of type modifyOBServers")
	} else if r.Spec.Type == constants.ClusterOpTypeSetParameters && r.Spec.SetParameters == nil {
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

	zoneReplicaMap := make(map[string]int)
	for _, z := range obcluster.Spec.Topology {
		zoneReplicaMap[z.Zone] = z.Replica
	}

	if r.Spec.AddZones != nil {
		for _, zone := range r.Spec.AddZones {
			if zone.Replica <= 0 {
				return nil, field.Invalid(field.NewPath("spec").Child("addZones").Child("replica"), zone.Replica, "replica must be greater than 0")
			}
			if _, ok := zoneReplicaMap[zone.Zone]; ok {
				return nil, field.Invalid(field.NewPath("spec").Child("addZones").Child("zone"), zone.Zone, "zone already exists")
			}
		}
	}

	if r.Spec.DeleteZones != nil {
		for _, zone := range r.Spec.DeleteZones {
			if _, ok := zoneReplicaMap[zone]; !ok {
				return nil, field.Invalid(field.NewPath("spec").Child("deleteZones").Child("zone"), zone, "zone does not exist")
			}
		}
	}

	if r.Spec.AdjustReplicas != nil {
		for _, alter := range r.Spec.AdjustReplicas {
			if alter.To <= 0 {
				return nil, field.Invalid(field.NewPath("spec").Child("adjustReplicas").Child("to"), alter.To, "replica must be greater than 0")
			}
			if alter.By < 0 {
				for _, zoneName := range alter.Zones {
					if zoneReplicaMap[zoneName]-alter.By <= 0 {
						return nil, field.Invalid(field.NewPath("spec").Child("adjustReplicas").Child("by"), alter.By, "number of replicas to scale down must be less than current replica")
					}
				}
			}
		}
	}

	if r.Spec.Type == constants.ClusterOpTypeModifyOBServers && r.Spec.ModifyOBServers != nil {
		modifySpec := r.Spec.ModifyOBServers
		if modifySpec.ExpandStorageSize != nil {
			if modifySpec.ExpandStorageSize.DataStorage.Cmp(obcluster.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 {
				return nil, field.Invalid(field.NewPath("spec").Child("expandStorageSize").Child("dataStorage"), modifySpec.ExpandStorageSize, "storage size can not be less than current size")
			}
			if modifySpec.ExpandStorageSize.LogStorage.Cmp(obcluster.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 {
				return nil, field.Invalid(field.NewPath("spec").Child("expandStorageSize").Child("logStorage"), modifySpec.ExpandStorageSize, "storage size can not be less than current size")
			}
			if modifySpec.ExpandStorageSize.RedoLogStorage.Cmp(obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0 {
				return nil, field.Invalid(field.NewPath("spec").Child("expandStorageSize").Child("redoLogStorage"), modifySpec.ExpandStorageSize, "storage size can not be less than current size")
			}
		} else if modifySpec.ModifyStorageClass != nil {
			if modifySpec.ModifyStorageClass.DataStorage != "" &&
				modifySpec.ModifyStorageClass.DataStorage != obcluster.Spec.OBServerTemplate.Storage.DataStorage.StorageClass &&
				validateStorageClassAllowExpansion(modifySpec.ModifyStorageClass.DataStorage) != nil {
				return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass").Child("dataStorage"), modifySpec.ModifyStorageClass, "storage class does not support expansion")
			}
			if modifySpec.ModifyStorageClass.LogStorage != "" &&
				modifySpec.ModifyStorageClass.LogStorage != obcluster.Spec.OBServerTemplate.Storage.LogStorage.StorageClass &&
				validateStorageClassAllowExpansion(modifySpec.ModifyStorageClass.LogStorage) != nil {
				return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass").Child("logStorage"), modifySpec.ModifyStorageClass, "storage class does not support expansion")
			}
			if modifySpec.ModifyStorageClass.RedoLogStorage != "" &&
				modifySpec.ModifyStorageClass.RedoLogStorage != obcluster.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass &&
				validateStorageClassAllowExpansion(modifySpec.ModifyStorageClass.RedoLogStorage) != nil {
				return nil, field.Invalid(field.NewPath("spec").Child("modifyStorageClass").Child("redoLogStorage"), modifySpec.ModifyStorageClass, "storage class does not support expansion")
			}
		}
	} else if r.Spec.Type == constants.ClusterOpTypeRestartOBServers && r.Spec.RestartOBServers != nil {
		if obcluster.Annotations[oceanbaseconst.AnnotationsSupportStaticIP] != "true" {
			return nil, field.Invalid(field.NewPath("spec").Child("obcluster"), r.Spec.OBCluster, "obcluster does not support static ip, can not restart observers")
		}
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
