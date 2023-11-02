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
	"fmt"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

func objectSentry(object any) {
	if object == nil {
		return
	}
	if cluster, ok := object.(*v1alpha1.OBCluster); ok {
		processOBCluster(cluster)
	} else if tenant, ok := object.(*v1alpha1.OBTenant); ok {
		processOBTenant(tenant)
	} else if server, ok := object.(*v1alpha1.OBServer); ok {
		processOBServer(server)
	} else if zone, ok := object.(*v1alpha1.OBZone); ok {
		processOBZone(zone)
	}
}

func processOBCluster(cluster *v1alpha1.OBCluster) {
	_, _ = fmt.Printf("[OBCluster Before] %+v\n", cluster)
	if cluster.Spec.BackupVolume != nil && cluster.Spec.BackupVolume.Volume != nil && cluster.Spec.BackupVolume.Volume.NFS != nil {
		cluster.Spec.BackupVolume.Volume.NFS.Server = md5Hash(cluster.Spec.BackupVolume.Volume.NFS.Server)
	}
	_, _ = fmt.Printf("[OBCluster After] %+v\n", cluster)
}

func processOBServer(server *v1alpha1.OBServer) {
	_, _ = fmt.Printf("[OBServer Before] %+v\n", server)
	server.Status.PodIp = md5Hash(server.Status.PodIp)
	server.Status.NodeIp = md5Hash(server.Status.NodeIp)
	_, _ = fmt.Printf("[OBServer After] %+v\n", server)
}

func processOBTenant(tenant *v1alpha1.OBTenant) {
	_, _ = fmt.Printf("[OBTenant After] %+v\n", tenant)
	for i := range tenant.Status.Pools {
		for j := range tenant.Status.Pools[i].Units {
			tenant.Status.Pools[i].Units[j].ServerIP = md5Hash(tenant.Status.Pools[i].Units[j].ServerIP)
			if tenant.Status.Pools[i].Units[j].Migrate.ServerIP != "" {
				tenant.Status.Pools[i].Units[j].Migrate.ServerIP = md5Hash(tenant.Status.Pools[i].Units[j].Migrate.ServerIP)
			}
		}
	}
	_, _ = fmt.Printf("[OBTenant After] %+v\n", tenant)
}

func processOBZone(zone *v1alpha1.OBZone) {
	_, _ = fmt.Printf("[OBZone Before] %+v\n", zone)
	for i := range zone.Status.OBServerStatus {
		zone.Status.OBServerStatus[i].Server = md5Hash(zone.Status.OBServerStatus[i].Server)
	}
	if zone.Spec.BackupVolume != nil && zone.Spec.BackupVolume.Volume != nil && zone.Spec.BackupVolume.Volume.NFS != nil {
		zone.Spec.BackupVolume.Volume.NFS.Server = md5Hash(zone.Spec.BackupVolume.Volume.NFS.Server)
	}
	_, _ = fmt.Printf("[OBZone After] %+v\n", zone)
}
