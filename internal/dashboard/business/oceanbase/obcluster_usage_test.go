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

package oceanbase

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func getMockedGVOBServers() []model.GVOBServer {
	return []model.GVOBServer{{
		ServerIP:          "1",
		Zone:              "zone1",
		CPUCapacity:       16,
		CPUAssigned:       3,
		MemCapacity:       16,
		MemAssigned:       3,
		MemoryLimit:       32,
		LogDiskCapacity:   100,
		LogDiskAssigned:   40,
		DataDiskCapacity:  100,
		DataDiskAllocated: 60,
	}, {
		ServerIP:          "2",
		Zone:              "zone1",
		CPUCapacity:       16,
		CPUAssigned:       10, // max in zone1
		MemCapacity:       16,
		MemAssigned:       3,
		MemoryLimit:       32,
		LogDiskCapacity:   100,
		LogDiskAssigned:   40,
		DataDiskCapacity:  100,
		DataDiskAllocated: 60,
	}, {
		ServerIP:          "3",
		Zone:              "zone1",
		CPUCapacity:       16,
		CPUAssigned:       3,
		MemCapacity:       16,
		MemAssigned:       12, // max in zone1
		MemoryLimit:       32,
		LogDiskCapacity:   100,
		LogDiskAssigned:   40,
		DataDiskCapacity:  100,
		DataDiskAllocated: 60,
	}, {
		ServerIP:          "4",
		Zone:              "zone2",
		CPUCapacity:       16,
		CPUAssigned:       3,
		MemCapacity:       16,
		MemAssigned:       5,
		MemoryLimit:       32,
		LogDiskCapacity:   100,
		LogDiskAssigned:   40,
		DataDiskCapacity:  100,
		DataDiskAllocated: 60,
	}}
}

var _ = Describe("Test OBClsuter usage", func() {
	It("Get observer usages", func() {
		servers, zoneMapping := getServerUsages(getMockedGVOBServers())
		Expect(servers).To(HaveLen(4))
		Expect(servers[0].AvailableCPU).To(Equal(int64(13)))
		Expect(servers[0].AvailableMemory).To(Equal(int64(13)))
		Expect(servers[1].AvailableCPU).To(Equal(int64(6)))
		Expect(servers[1].AvailableMemory).To(Equal(int64(13)))
		Expect(servers[2].AvailableCPU).To(Equal(int64(13)))
		Expect(servers[2].AvailableMemory).To(Equal(int64(4)))

		Expect(zoneMapping).To(HaveLen(2))
		Expect(zoneMapping).To(HaveKey("zone1"))
		Expect(zoneMapping).To(HaveKey("zone2"))
		Expect(zoneMapping["zone1"].AvailableCPU).To(Equal(int64(6)))
		Expect(zoneMapping["zone1"].AvailableMemory).To(Equal(int64(4)))
		Expect(zoneMapping["zone1"].AvailableDataDisk).To(Equal(int64(40)))
		Expect(zoneMapping["zone1"].AvailableLogDisk).To(Equal(int64(60)))
		Expect(zoneMapping["zone2"].AvailableCPU).To(Equal(int64(13)))
		Expect(zoneMapping["zone2"].AvailableMemory).To(Equal(int64(11)))
		Expect(zoneMapping["zone2"].AvailableDataDisk).To(Equal(int64(40)))
		Expect(zoneMapping["zone2"].AvailableLogDisk).To(Equal(int64(60)))
	})
})
