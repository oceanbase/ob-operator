package obzone

import "github.com/oceanbase/ob-operator/api/v1alpha1"

func (m *OBZoneManager) doStorageSizeExpand(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 ||
		observer.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 ||
		observer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0
}

func (m *OBZoneManager) doCalcResourceChange(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		observer.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Memory) != 0
}
