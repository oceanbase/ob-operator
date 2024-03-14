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
	"context"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/oceanbase"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
)

func CreateTenantPool(ctx context.Context, nn param.TenantPoolName, p *param.TenantPoolSpec) (bool, error) {
	cpuCount, err := resource.ParseQuantity(p.UnitConfig.CPUCount)
	if err != nil {
		return false, oberr.NewBadRequest("invalid cpu count: " + err.Error())
	}
	memorySize, err := resource.ParseQuantity(p.UnitConfig.MemorySize)
	if err != nil {
		return false, oberr.NewBadRequest("invalid memory size: " + err.Error())
	}
	logDiskSize, err := resource.ParseQuantity(p.UnitConfig.LogDiskSize)
	if err != nil {
		return false, oberr.NewBadRequest("invalid log disk size: " + err.Error())
	}

	tenantCR, err := oceanbase.GetOBTenant(ctx, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		return false, err
	}
	for _, pool := range tenantCR.Spec.Pools {
		if pool.Zone == nn.ZoneName {
			return false, oberr.NewBadRequest("pool already exists")
		}
	}
	clusterCR, err := oceanbase.GetOBCluster(ctx, nn.Namespace, tenantCR.Spec.ClusterName)
	if err != nil {
		return false, err
	}
	for _, zone := range clusterCR.Spec.Topology {
		if zone.Zone == nn.ZoneName {
			tenantCR.Spec.Pools = append(tenantCR.Spec.Pools, v1alpha1.ResourcePoolSpec{
				Zone:     nn.ZoneName,
				Priority: p.Priority,
				Type: &v1alpha1.LocalityType{
					Name:     "Full",
					Replica:  1,
					IsActive: true,
				},
				UnitConfig: &v1alpha1.UnitConfig{
					MaxCPU:      cpuCount,
					MemorySize:  memorySize,
					MinCPU:      cpuCount,
					MaxIops:     p.UnitConfig.MaxIops,
					MinIops:     p.UnitConfig.MaxIops,
					IopsWeight:  p.UnitConfig.IopsWeight,
					LogDiskSize: logDiskSize,
				},
			})
			_, err = oceanbase.UpdateOBTenant(ctx, tenantCR)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, oberr.NewBadRequest("zone not found in the cluster")
}

func DeleteTenantPool(ctx context.Context, nn param.TenantPoolName) (bool, error) {
	tenantCR, err := oceanbase.GetOBTenant(ctx, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		return false, err
	}
	if len(tenantCR.Spec.Pools) == 1 {
		return false, oberr.NewBadRequest("at least one pool is required")
	}
	remainPools := make([]v1alpha1.ResourcePoolSpec, 0, len(tenantCR.Spec.Pools)-1)
	for i, pool := range tenantCR.Spec.Pools {
		if pool.Zone != nn.ZoneName {
			remainPools = append(remainPools, tenantCR.Spec.Pools[i])
		}
	}
	if len(remainPools) == len(tenantCR.Spec.Pools) {
		return false, oberr.NewBadRequest("pool not found")
	}

	tenantCR.Spec.Pools = remainPools
	_, err = oceanbase.UpdateOBTenant(ctx, tenantCR)
	if err != nil {
		return false, err
	}
	return true, nil
}

func PatchTenantPool(ctx context.Context, nn param.TenantPoolName, p *param.TenantPoolSpec) (bool, error) {
	tenantCR, err := oceanbase.GetOBTenant(ctx, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		return false, err
	}
	for i, pool := range tenantCR.Spec.Pools {
		if pool.Zone == nn.ZoneName {
			tenantCR.Spec.Pools[i].Priority = p.Priority
			if tenantCR.Spec.Pools[i].UnitConfig != nil {
				if cpuCount, err := resource.ParseQuantity(p.UnitConfig.CPUCount); err == nil {
					tenantCR.Spec.Pools[i].UnitConfig.MaxCPU = cpuCount
					tenantCR.Spec.Pools[i].UnitConfig.MinCPU = cpuCount
				}
				if memorySize, err := resource.ParseQuantity(p.UnitConfig.MemorySize); err == nil {
					tenantCR.Spec.Pools[i].UnitConfig.MemorySize = memorySize
				}
				if logDiskSize, err := resource.ParseQuantity(p.UnitConfig.LogDiskSize); err == nil {
					tenantCR.Spec.Pools[i].UnitConfig.LogDiskSize = logDiskSize
				}
				if p.UnitConfig.MaxIops != 0 {
					tenantCR.Spec.Pools[i].UnitConfig.MaxIops = p.UnitConfig.MaxIops
				}
				if p.UnitConfig.MinIops != 0 {
					tenantCR.Spec.Pools[i].UnitConfig.MinIops = p.UnitConfig.MinIops
				}
				if p.UnitConfig.IopsWeight != 0 {
					tenantCR.Spec.Pools[i].UnitConfig.IopsWeight = p.UnitConfig.IopsWeight
				}
			}
			_, err = oceanbase.UpdateOBTenant(ctx, tenantCR)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, oberr.NewBadRequest("pool not found")
}
