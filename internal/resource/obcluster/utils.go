package obcluster

import "github.com/oceanbase/ob-operator/api/v1alpha1"

func (m *OBClusterManager) doStorageSizeExpand(obzone *v1alpha1.OBZone) bool {
	return obzone.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(m.OBCluster.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 ||
		obzone.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(m.OBCluster.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 ||
		obzone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(m.OBCluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0
}

func (m *OBClusterManager) doCalcResourceChange(obzone *v1alpha1.OBZone) bool {
	return obzone.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBCluster.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		obzone.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBCluster.Spec.OBServerTemplate.Resource.Memory) != 0
}
