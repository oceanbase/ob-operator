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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func (r *OBLogServiceCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
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
		if topo.Region == "" {
			return nil, fmt.Errorf("topology[%d].region is required", i)
		}
		if topo.Replica < 1 {
			return nil, fmt.Errorf("topology[%d].replica must be at least 1", i)
		}
	}
	if r.Spec.ClusterId <= 0 {
		return nil, errors.New("clusterId must be a positive integer")
	}
	if r.Spec.Image == "" {
		return nil, errors.New("image is required")
	}
	if r.Spec.ObjectStoreURL.BucketURL == "" {
		return nil, errors.New("objectStoreUrl.bucketURL is required")
	}
	if r.Spec.ObjectStoreURL.SecretRef.Name == "" {
		return nil, errors.New("objectStoreUrl.secretRef.name is required")
	}
	if r.Spec.Storage == nil {
		return nil, errors.New("storage is required")
	}
	return nil, nil
}

func (r *OBLogServiceCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldLS, ok := old.(*OBLogServiceCluster)
	if !ok {
		return nil, errors.New("failed to convert old object to OBLogServiceCluster")
	}
	if oldLS.Spec.Image != r.Spec.Image {
		return nil, errors.New("OBLogServiceCluster does not support image upgrade in this version")
	}
	if oldLS.Spec.ClusterId != r.Spec.ClusterId {
		return nil, errors.New("clusterId cannot be changed after creation")
	}
	if !reflect.DeepEqual(oldLS.Spec.ObjectStoreURL, r.Spec.ObjectStoreURL) {
		return nil, errors.New("objectStoreUrl cannot be changed after creation")
	}
	if !reflect.DeepEqual(oldLS.Spec.Storage, r.Spec.Storage) {
		return nil, errors.New("storage cannot be changed after creation")
	}
	return nil, nil
}

func (r *OBLogServiceCluster) ValidateDelete() (admission.Warnings, error) {
	if r.Annotations[oceanbaseconst.AnnotationsIgnoreDeletion] == "true" {
		return nil, apierrors.NewBadRequest("OBLogServiceCluster " + r.Name + " is protected from deletion by annotation " + oceanbaseconst.AnnotationsIgnoreDeletion)
	}
	return nil, nil
}
