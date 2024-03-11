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

package telemetry

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/telemetry/models"
)

func TransformReportOBCluster(c *v1alpha1.OBCluster) *models.OBCluster {
	res := &models.OBCluster{
		ClusterName:            c.Spec.ClusterName,
		ClusterId:              c.Spec.ClusterId,
		ClusterMode:            "",
		Image:                  "",
		CPU:                    0,
		Memory:                 0,
		SysLogStorage:          &models.StorageSpec{},
		DataStorage:            &models.StorageSpec{},
		RedoLogStorage:         &models.StorageSpec{},
		ConfiguredBackupVolume: c.Spec.BackupVolume != nil,
		ConfiguredMonitor:      c.Spec.MonitorTemplate != nil,
		Zones:                  []models.OBZoneStatus{},
		CommonFields: models.CommonFields{
			Name:      c.GetObjectMeta().GetName(),
			Namespace: c.GetObjectMeta().GetNamespace(),
			UID:       string(c.GetObjectMeta().GetUID()),
			Status:    c.Status.Status,
		},
	}
	if ca := c.GetAnnotations(); ca != nil {
		if v, ok := ca[oceanbaseconst.AnnotationsMode]; ok {
			res.ClusterMode = v
		}
		if _, ok := ca[oceanbaseconst.AnnotationsSinglePVC]; ok {
			res.SinglePVC = true
		}
		if _, ok := ca[oceanbaseconst.AnnotationsIndependentPVCLifecycle]; ok {
			res.IndependentPVC = true
		}
	}
	if c.Spec.OBServerTemplate != nil {
		res.Image = c.Spec.OBServerTemplate.Image
		res.CPU = c.Spec.OBServerTemplate.Resource.Cpu.Value()
		res.Memory = c.Spec.OBServerTemplate.Resource.Memory.Value()
		res.SysLogStorage = &models.StorageSpec{
			StorageClass: c.Spec.OBServerTemplate.Storage.LogStorage.StorageClass,
			StorageSize:  c.Spec.OBServerTemplate.Storage.LogStorage.Size.Value(),
		}
		res.DataStorage = &models.StorageSpec{
			StorageClass: c.Spec.OBServerTemplate.Storage.DataStorage.StorageClass,
			StorageSize:  c.Spec.OBServerTemplate.Storage.DataStorage.Size.Value(),
		}
		res.RedoLogStorage = &models.StorageSpec{
			StorageClass: c.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass,
			StorageSize:  c.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Value(),
		}
	}
	replicaMapping := make(map[string]int, len(c.Status.OBZoneStatus))
	for _, z := range c.Spec.Topology {
		replicaMapping[z.Zone] = z.Replica
	}
	for _, z := range c.Status.OBZoneStatus {
		if replica, ok := replicaMapping[z.Zone]; ok {
			res.Zones = append(res.Zones, models.OBZoneStatus{
				ZoneName: z.Zone,
				Replica:  replica,
				Status:   z.Status,
			})
		}
	}
	if c.Status.OperationContext != nil {
		res.RunningFlow = string(c.Status.OperationContext.Name)
		res.RunningTask = string(c.Status.OperationContext.Task)
		res.TaskStatus = string(c.Status.OperationContext.TaskStatus)
	}
	return res
}

func TransformReportOBZone(c *v1alpha1.OBZone) *models.OBZone {
	res := &models.OBZone{
		ClusterName: c.Spec.ClusterName,
		ClusterId:   c.Spec.ClusterId,
		ClusterCR:   "",
		Image:       "",
		CommonFields: models.CommonFields{
			Status:    c.Status.Status,
			Name:      c.GetObjectMeta().GetName(),
			Namespace: c.GetObjectMeta().GetNamespace(),
			UID:       string(c.GetObjectMeta().GetUID()),
		},
	}
	if c.Spec.OBServerTemplate != nil {
		res.Image = c.Spec.OBServerTemplate.Image
	}
	if anno := c.GetAnnotations(); anno != nil {
		if v, ok := anno[oceanbaseconst.LabelRefOBCluster]; ok {
			res.ClusterCR = v
		}
	}
	if c.Status.OperationContext != nil {
		res.RunningFlow = string(c.Status.OperationContext.Name)
		res.RunningTask = string(c.Status.OperationContext.Task)
		res.TaskStatus = string(c.Status.OperationContext.TaskStatus)
	}
	return res
}

func TransformReportOBServer(c *v1alpha1.OBServer) *models.OBServer {
	res := &models.OBServer{
		ClusterName:   c.Spec.ClusterName,
		ClusterId:     c.Spec.ClusterId,
		ClusterCR:     "",
		ZoneName:      c.Spec.Zone,
		Image:         "",
		CNI:           c.Status.CNI,
		PodPhase:      string(c.Status.PodPhase),
		PodIPHash:     md5Hash(c.Status.PodIp),
		ServiceIPHash: md5Hash(c.Status.ServiceIp),
		CommonFields: models.CommonFields{
			Status:    c.Status.Status,
			Name:      c.GetObjectMeta().GetName(),
			Namespace: c.GetObjectMeta().GetNamespace(),
			UID:       string(c.GetObjectMeta().GetUID()),
		},
	}
	if c.Spec.OBServerTemplate != nil {
		res.Image = c.Spec.OBServerTemplate.Image
	}
	if anno := c.GetAnnotations(); anno != nil {
		if v, ok := anno[oceanbaseconst.LabelRefOBCluster]; ok {
			res.ClusterCR = v
		}
	}
	if c.Status.OperationContext != nil {
		res.RunningFlow = string(c.Status.OperationContext.Name)
		res.RunningTask = string(c.Status.OperationContext.Task)
		res.TaskStatus = string(c.Status.OperationContext.TaskStatus)
	}
	return res
}

func TransformReportOBTenant(c *v1alpha1.OBTenant) *models.OBTenant {
	res := &models.OBTenant{
		TenantName:             c.Spec.TenantName,
		ClusterName:            c.Spec.ClusterName,
		TenantRole:             string(c.Spec.TenantRole),
		UnitNumber:             c.Spec.UnitNumber,
		PrimaryTenant:          "",
		RestoreArchiveDestType: "",
		RestoreBakDataDestType: "",
		Topology:               []models.OBTenantResourcePool{},
		CommonFields: models.CommonFields{
			Status:    c.Status.Status,
			Name:      c.GetObjectMeta().GetName(),
			Namespace: c.GetObjectMeta().GetNamespace(),
			UID:       string(c.GetObjectMeta().GetUID()),
		},
	}
	if c.Status.OperationContext != nil {
		res.RunningFlow = string(c.Status.OperationContext.Name)
		res.RunningTask = string(c.Status.OperationContext.Task)
		res.TaskStatus = string(c.Status.OperationContext.TaskStatus)
	}
	if c.Spec.Source != nil {
		if c.Spec.Source.Tenant != nil {
			res.PrimaryTenant = *c.Spec.Source.Tenant
		}
		if c.Spec.Source.Restore != nil {
			res.RestoreArchiveDestType = string(c.Spec.Source.Restore.ArchiveSource.Type)
			res.RestoreBakDataDestType = string(c.Spec.Source.Restore.BakDataSource.Type)
		}
	}
	for _, p := range c.Status.Pools {
		res.Topology = append(res.Topology, models.OBTenantResourcePool{
			Zone:        p.ZoneList,
			Priority:    p.Priority,
			Type:        p.Type.Name,
			MaxCPU:      p.UnitConfig.MaxCPU.Value(),
			MinCPU:      p.UnitConfig.MinCPU.Value(),
			MemorySize:  p.UnitConfig.MemorySize.Value(),
			MaxIOPS:     p.UnitConfig.MaxIops,
			MinIOPS:     p.UnitConfig.MinIops,
			IOPSWeight:  p.UnitConfig.IopsWeight,
			LogDiskSize: p.UnitConfig.LogDiskSize.Value(),
			UnitNumber:  p.UnitNumber,
		})
	}
	return res
}

func TransformReportOBBackupPolicy(c *v1alpha1.OBTenantBackupPolicy) *models.OBBackupPolicy {
	res := &models.OBBackupPolicy{
		TenantCR:                   c.Spec.TenantCRName,
		TenantName:                 c.Spec.TenantName,
		ArchiveDestType:            string(c.Spec.LogArchive.Destination.Type),
		ArchiveSwitchPieceInterval: string(c.Spec.LogArchive.SwitchPieceInterval),
		BakDataDestType:            string(c.Spec.DataBackup.Destination.Type),
		BakDataFullCrontab:         c.Spec.DataBackup.FullCrontab,
		BakDataIncrCrontab:         c.Spec.DataBackup.IncrementalCrontab,
		EncryptBakData:             c.Spec.DataBackup.EncryptionSecret != "",
		CommonFields: models.CommonFields{
			Status:    string(c.Status.Status),
			Name:      c.GetObjectMeta().GetName(),
			Namespace: c.GetObjectMeta().GetNamespace(),
			UID:       string(c.GetObjectMeta().GetUID()),
		},
	}
	if c.Status.OperationContext != nil {
		res.RunningFlow = string(c.Status.OperationContext.Name)
		res.RunningTask = string(c.Status.OperationContext.Task)
		res.TaskStatus = string(c.Status.OperationContext.TaskStatus)
	}
	return res
}
