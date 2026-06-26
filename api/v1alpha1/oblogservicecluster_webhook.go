/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1alpha1

import (
	"errors"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

var oblogserviceclusterlog = logf.Log.WithName("oblogservicecluster-resource")

func (r *OBLogServiceCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-oceanbase-oceanbase-com-v1alpha1-oblogservicecluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=oblogserviceclusters,verbs=create;update,versions=v1alpha1,name=moblogservicecluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OBLogServiceCluster{}

func (r *OBLogServiceCluster) Default() {
	logger := oblogserviceclusterlog.WithValues("namespace", r.Namespace, "name", r.Name)

	parameterMap := make(map[string]apitypes.Parameter)
	if r.Spec.LogService == nil || r.Spec.LogService.Resource.Memory.IsZero() {
		logger.Error(errors.New("logService.resource.memory is missing"), "parse logservice's memory size failed")
	} else {
		memorySize, ok := r.Spec.LogService.Resource.Memory.AsInt64()
		if ok {
			memoryLimit := fmt.Sprintf("%dM", memorySize*int64(obcfg.GetConfig().Resource.DefaultMemoryLimitPercent)/100/int64(oceanbaseconst.MegaConverter))
			parameterMap["memory_limit"] = apitypes.Parameter{
				Name:  "memory_limit",
				Value: memoryLimit,
			}
		} else {
			logger.Error(errors.New("failed to parse memory size"), "parse logservice's memory size failed")
		}
	}

	for _, parameter := range r.Spec.Parameters {
		parameterMap[parameter.Name] = parameter
	}
	parameters := make([]apitypes.Parameter, 0, len(parameterMap))
	for _, v := range parameterMap {
		parameters = append(parameters, v)
	}
	r.Spec.Parameters = parameters
}

//+kubebuilder:webhook:path=/validate-oceanbase-oceanbase-com-v1alpha1-oblogservicecluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=oceanbase.oceanbase.com,resources=oblogserviceclusters,verbs=create;update;delete,versions=v1alpha1,name=voblogservicecluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OBLogServiceCluster{}

func (r *OBLogServiceCluster) ValidateCreate() (admission.Warnings, error) {
	if len(r.Spec.Topology) < 1 {
		return nil, errors.New("topology must have at least 1 zone")
	}
	seenZones := make(map[string]bool)
	for i, topo := range r.Spec.Topology {
		if topo.Zone == "" {
			return nil, fmt.Errorf("topology[%d].zone is required", i)
		}
		if seenZones[topo.Zone] {
			return nil, fmt.Errorf("topology[%d].zone %q is duplicated", i, topo.Zone)
		}
		seenZones[topo.Zone] = true
		if topo.Replica < 1 {
			return nil, fmt.Errorf("topology[%d].replica must be at least 1", i)
		}
	}
	if r.Spec.ClusterId <= 0 {
		return nil, errors.New("clusterId must be a positive integer")
	}
	if r.Spec.LogService == nil {
		return nil, errors.New("logService is required")
	}
	if r.Spec.LogService.Image == "" {
		return nil, errors.New("logService.image is required")
	}
	if r.Spec.ObjectStoreURL.BucketURL == "" {
		return nil, errors.New("objectStoreUrl.bucketURL is required")
	}
	if r.Spec.ObjectStoreURL.SecretRef.Name == "" {
		return nil, errors.New("objectStoreUrl.secretRef.name is required")
	}
	if r.Spec.LogService.Storage == nil {
		return nil, errors.New("logService.storage is required")
	}
	if r.Spec.LogService.Resource.Memory.IsZero() {
		return nil, errors.New("logService.resource.memory is required")
	}
	for _, p := range r.Spec.Parameters {
		if p.Name == "memory_limit" {
			memoryLimit, err := resource.ParseQuantity(p.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse memory_limit parameter: %w", err)
			}
			if memoryLimit.AsApproximateFloat64() > r.Spec.LogService.Resource.Memory.AsApproximateFloat64() {
				return nil, errors.New("memory_limit exceeds logService.resource.memory")
			}
			break
		}
	}
	return nil, nil
}

func (r *OBLogServiceCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldLS, ok := old.(*OBLogServiceCluster)
	if !ok {
		return nil, errors.New("failed to convert old object to OBLogServiceCluster")
	}
	if oldLS.Spec.ClusterId != r.Spec.ClusterId {
		return nil, errors.New("clusterId cannot be changed after creation")
	}
	if !reflect.DeepEqual(oldLS.Spec.ObjectStoreURL, r.Spec.ObjectStoreURL) {
		return nil, errors.New("objectStoreUrl cannot be changed after creation")
	}
	// The entire logService template (image/resource/storage) is immutable. Only
	// topology[].replica is mutable (handled by the ModifyZoneReplica flow); there
	// is no flow that reconciles resource changes, so allowing them would be a
	// silent no-op.
	if !reflect.DeepEqual(oldLS.Spec.LogService, r.Spec.LogService) {
		return nil, errors.New("logService template cannot be changed after creation")
	}
	oldZones := make(map[string]bool, len(oldLS.Spec.Topology))
	for _, t := range oldLS.Spec.Topology {
		oldZones[t.Zone] = true
	}
	newZones := make(map[string]bool, len(r.Spec.Topology))
	for _, t := range r.Spec.Topology {
		newZones[t.Zone] = true
	}
	for z := range newZones {
		if !oldZones[z] {
			return nil, errors.New("adding zones is not supported for OBLogServiceCluster")
		}
	}
	for z := range oldZones {
		if !newZones[z] {
			return nil, errors.New("removing zones is not supported for OBLogServiceCluster")
		}
	}
	return nil, nil
}

func (r *OBLogServiceCluster) ValidateDelete() (admission.Warnings, error) {
	if r.Annotations[oceanbaseconst.AnnotationsIgnoreDeletion] == "true" {
		return nil, apierrors.NewBadRequest("OBLogServiceCluster " + r.Name + " is protected from deletion by annotation " + oceanbaseconst.AnnotationsIgnoreDeletion)
	}
	return nil, nil
}
