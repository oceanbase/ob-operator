package utils

import (
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

// EffectiveZoneTemplate is the template actually copied to a newly created Observer.
func EffectiveZoneTemplate(zone *v1alpha1.OBZone) *apitypes.OBServerTemplate {
	if zone.Spec.Topology.OBServerTemplate != nil {
		return zone.Spec.Topology.OBServerTemplate
	}
	return zone.Spec.OBServerTemplate
}

func SameObserverStorageClasses(a, b *apitypes.OBServerTemplate, annotations map[string]string) bool {
	if a == nil || b == nil || a.Storage == nil || b.Storage == nil {
		return false
	}
	classEqual := func(x, y *apitypes.StorageSpec) bool {
		return x != nil && y != nil && x.StorageClass == y.StorageClass
	}
	x, y := a.Storage, b.Storage
	if !classEqual(x.DataStorage, y.DataStorage) {
		return false
	}
	single, explicit := annotations[oceanbaseconst.AnnotationsSinglePVC]
	if single == "true" {
		return true
	} // the one PVC uses the data storage class
	if !classEqual(x.LogStorage, y.LogStorage) {
		return false
	}
	if annotations[oceanbaseconst.AnnotationsDeploymentMode] == oceanbaseconst.DeploymentModeSharedStorage {
		return true
	}
	if x.RedoLogStorage == nil || y.RedoLogStorage == nil {
		// Old persisted objects without a layout annotation may have omitted redo.
		return !explicit || (x.RedoLogStorage == nil) == (y.RedoLogStorage == nil)
	}
	return classEqual(x.RedoLogStorage, y.RedoLogStorage)
}
