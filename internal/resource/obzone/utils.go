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

package obzone

import "github.com/oceanbase/ob-operator/api/v1alpha1"

func (m *OBZoneManager) checkIfStorageSizeExpand(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 ||
		observer.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 ||
		observer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0
}

func (m *OBZoneManager) checkIfCalcResourceChange(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		observer.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Memory) != 0
}
