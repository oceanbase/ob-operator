/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

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
