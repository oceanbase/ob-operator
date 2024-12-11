/*
Copyright (c) 2024 OceanBase
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
	"fmt"
	"math"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"

	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

// MapZonesToTopology map --zones to zoneTopology
func MapZonesToTopology(zones map[string]string) ([]param.ZoneTopology, error) {
	if zones == nil {
		return nil, fmt.Errorf("zone replica is required")
	}
	topology := make([]param.ZoneTopology, 0)
	for zoneName, replicaStr := range zones {
		replica, err := strconv.Atoi(replicaStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for zone %s: %s", zoneName, replicaStr)
		}
		topology = append(topology, param.ZoneTopology{
			Zone:         zoneName,
			Replicas:     replica,
			NodeSelector: make([]common.KVPair, 0),
			Affinities:   make([]common.AffinitySpec, 0),
		})
	}
	return topology, nil
}

// MapZonesToPools map --zones to []resourcePool
func MapZonesToPools(zones map[string]string) ([]param.ResourcePoolSpec, error) {
	if zones == nil {
		return nil, fmt.Errorf("Zone priority is required")
	}
	resourcePool := make([]param.ResourcePoolSpec, 0)
	for zoneName, priorityStr := range zones {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for zone %s: %s", zoneName, priorityStr)
		}
		resourcePool = append(resourcePool, param.ResourcePoolSpec{
			Zone:     zoneName,
			Priority: priority,
			Type:     "Full",
		})
	}
	return resourcePool, nil
}

// MapParameters map --parameters to parameters
func MapParameters(parameters map[string]string) ([]common.KVPair, error) {
	kvMap := make([]common.KVPair, 0)
	for k, v := range parameters {
		kvMap = append(kvMap, common.KVPair{
			Key:   k,
			Value: v,
		})
	}
	return kvMap, nil
}

// ParseUnitConfig parse param.UnitConfig to v1alpha1.UnitConfig
func ParseUnitConfig(unitConfig *param.UnitConfig) (*v1alpha1.UnitConfig, error) {
	cpuCount, err := resource.ParseQuantity(unitConfig.CPUCount)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid cpu count: " + err.Error())
	}
	memorySize, err := resource.ParseQuantity(unitConfig.MemorySize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid memory size: " + err.Error())
	}
	logDiskSize, err := resource.ParseQuantity(unitConfig.LogDiskSize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid log disk size: " + err.Error())
	}
	var maxIops, minIops int
	if unitConfig.MaxIops > math.MaxInt32 {
		maxIops = math.MaxInt32
	} else {
		maxIops = int(unitConfig.MaxIops)
	}
	if unitConfig.MinIops > math.MaxInt32 {
		minIops = math.MaxInt32
	} else {
		minIops = int(unitConfig.MinIops)
	}
	return &v1alpha1.UnitConfig{
		MaxCPU:      cpuCount,
		MemorySize:  memorySize,
		MinCPU:      cpuCount,
		LogDiskSize: logDiskSize,
		MaxIops:     maxIops,
		MinIops:     minIops,
		IopsWeight:  unitConfig.IopsWeight,
	}, nil
}
