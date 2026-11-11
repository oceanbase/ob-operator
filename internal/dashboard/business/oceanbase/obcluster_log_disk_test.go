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
	"encoding/json"
	"math"
	"testing"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func TestServerLogDiskCapacity(t *testing.T) {
	const gib int64 = 1 << 30
	for _, tt := range []struct {
		name                          string
		capacity, assigned, available int64
		unlimited                     bool
	}{
		{"shared log service", math.MaxInt64 - 1, 7 * gib, 0, true},
		{"shared log service without allocations", math.MaxInt64 - 1, 0, 0, true},
		{"finite capacity", 20 * gib, 7 * gib, 13 * gib, false},
		{"exhausted capacity", 7 * gib, 7 * gib, 0, false},
		{"over allocated", 6 * gib, 7 * gib, 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			servers, zones := getServerUsages([]model.GVOBServer{{ServerIP: "server1", Zone: "zone1", LogDiskCapacity: tt.capacity, LogDiskAssigned: tt.assigned}})
			for name, value := range map[string]interface{}{"server": servers[0], "zone": zones["zone1"]} {
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var got struct {
					Available int64 `json:"availableLogDisk"`
					Unlimited bool  `json:"logDiskUnlimited"`
				}
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatal(err)
				}
				if got.Available != tt.available || got.Unlimited != tt.unlimited {
					t.Errorf("%s: got available=%d unlimited=%v; want available=%d unlimited=%v", name, got.Available, got.Unlimited, tt.available, tt.unlimited)
				}
			}
		})
	}
}

func TestZoneLogDiskCapacityMixedServers(t *testing.T) {
	const gib int64 = 1 << 30
	for _, reverse := range []bool{false, true} {
		input := []model.GVOBServer{
			{ServerIP: "finite", Zone: "zone1", LogDiskCapacity: 20 * gib, LogDiskAssigned: 7 * gib},
			{ServerIP: "unlimited", Zone: "zone1", LogDiskCapacity: math.MaxInt64 - 1, LogDiskAssigned: 7 * gib},
		}
		if reverse {
			input[0], input[1] = input[1], input[0]
		}
		servers, zones := getServerUsages(input)
		data, err := json.Marshal(zones["zone1"])
		if err != nil {
			t.Fatal(err)
		}
		var got struct {
			Available int64 `json:"availableLogDisk"`
			Unlimited bool  `json:"logDiskUnlimited"`
		}
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if !got.Unlimited || got.Available != 0 {
			t.Errorf("reverse=%v: zone should allow placement on unlimited server: %s", reverse, data)
		}
		for i, s := range servers {
			if input[i].ServerIP == "finite" && s.AvailableLogDisk != 13*gib {
				t.Errorf("finite server was modified by zone aggregation: %+v", s)
			}
		}
	}
}
