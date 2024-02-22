/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obcluster

import "github.com/oceanbase/ob-operator/api/v1alpha1"

func (m *OBClusterManager) checkIfStorageSizeExpand(obzone *v1alpha1.OBZone) bool {
	return obzone.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(m.OBCluster.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 ||
		obzone.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(m.OBCluster.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 ||
		obzone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(m.OBCluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0
}

func (m *OBClusterManager) checkIfCalcResourceChange(obzone *v1alpha1.OBZone) bool {
	return obzone.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBCluster.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		obzone.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBCluster.Spec.OBServerTemplate.Resource.Memory) != 0
}
