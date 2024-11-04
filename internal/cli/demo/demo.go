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
package demo

import (
	"errors"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/internal/cli/cluster"
	"github.com/oceanbase/ob-operator/internal/cli/tenant"
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

func SetDefaultClusterConf(clusterType string, o *cluster.CreateOptions) error {
	o.BackupVolume = nil // set to nil to disable backup volume
	o.Parameters = []modelcommon.KVPair{
		{
			Key:   cluster.FLAG_MIN_FULL_RESOURCE_POOL_MEMORY,
			Value: cluster.DEFAULT_MIN_FULL_RESOURCE_POOL_MEMORY,
		},
		{
			Key:   cluster.FLAG_SYSTEM_MEMORY,
			Value: cluster.DEFAULT_SYSTEM_MEMORY,
		},
	}
	o.OBServer = &param.OBServerSpec{
		Image: cluster.DEFAULT_OBSERVER_IMAGE,
		Resource: modelcommon.ResourceSpec{
			Cpu:      cluster.DEFAULT_OBSERVER_CPU,
			MemoryGB: cluster.DEFAULT_OBSERVER_MEMORY,
		},
		Storage: &param.OBServerStorageSpec{
			Data: modelcommon.StorageSpec{
				StorageClass: cluster.DEFAULT_DATA_STORAGE_CLASS,
				SizeGB:       cluster.DEFAULT_DATA_STORAGE_SIZE,
			},
			RedoLog: modelcommon.StorageSpec{
				StorageClass: cluster.DEFAULT_REDO_LOG_STORAGE_CLASS,
				SizeGB:       cluster.DEFAULT_REDO_LOG_STORAGE_SIZE,
			},
			Log: modelcommon.StorageSpec{
				StorageClass: cluster.DEFAULT_LOG_STORAGE_CLASS,
				SizeGB:       cluster.DEFAULT_LOG_STORAGE_SIZE,
			},
		},
	}
	o.Monitor = &param.MonitorSpec{
		Image: cluster.DEFAULT_MONITOR_IMAGE,
		Resource: modelcommon.ResourceSpec{
			Cpu:      cluster.DEFAULT_MONITOR_CPU,
			MemoryGB: cluster.DEFAULT_MONITOR_MEMORY,
		},
	}
	switch clusterType {
	case cluster.SINGLE_NODE:
		o.Topology = []param.ZoneTopology{
			{
				Zone:     "z1",
				Replicas: 1,
			},
		}
	case cluster.THREE_NODE:
		o.Topology = []param.ZoneTopology{
			{
				Zone:     "z1",
				Replicas: 1,
			},
			{
				Zone:     "z2",
				Replicas: 1,
			},
			{
				Zone:     "z3",
				Replicas: 1,
			},
		}
	default:
		return errors.New(cluster.ErrInvalidClusterType)
	}
	return nil
}

func SetDefaultTenantConf(clusterType string, namespace string, clusterName string, o *tenant.CreateOptions) error {
	o.Namespace = namespace
	o.ClusterName = clusterName
	o.TenantRole = string(apiconst.TenantRolePrimary)
	o.Charset = tenant.DEFAULT_CHARSET
	o.ConnectWhiteList = tenant.DEFAULT_CONNECT_WHITE_LIST
	o.UnitNumber = tenant.DEFAULT_UNIT_NUMBER
	o.Source = nil // set to nil to avoid creating tenant with source
	o.UnitConfig = &param.UnitConfig{
		MaxIops:     tenant.DEFAULT_MAX_IOPS,
		MinIops:     tenant.DEFAULT_MIN_IOPS,
		IopsWeight:  tenant.DEFAULT_IOPS_WEIGHT,
		CPUCount:    tenant.DEFAULT_CPU_COUNT,
		MemorySize:  tenant.DEFAULT_MEMORY_SIZE,
		LogDiskSize: tenant.DEFAULT_LOG_DISK_SIZE,
	}
	switch clusterType {
	case cluster.SINGLE_NODE:
		o.Pools = []param.ResourcePoolSpec{
			{
				Zone:     "z1",
				Priority: 1,
				Type:     "full",
			},
		}
	case cluster.THREE_NODE:
		o.Pools = []param.ResourcePoolSpec{
			{
				Zone:     "z1",
				Priority: 1,
				Type:     "full",
			},
			{
				Zone:     "z2",
				Priority: 1,
				Type:     "full",
			},
			{
				Zone:     "z3",
				Priority: 1,
				Type:     "full",
			},
		}
	default:
		return errors.New(cluster.ErrInvalidClusterType)
	}
	return nil
}
