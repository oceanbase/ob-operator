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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

// Targets of sentry:
// 1. Digests IP addresses of servers
// 2. Digests NFS server address of backup volume
// 3. Remove redundant fields in status

func objectSentry(object any) {
	if object == nil {
		return
	}

	if metaObj, ok := object.(metav1.Object); ok {
		// remove managed fields which are of no interest
		metaObj.SetManagedFields(nil)
	}

	if cluster, ok := object.(*v1alpha1.OBCluster); ok {
		debugWrapper(processOBCluster, cluster, "OBCluster")
	} else if tenant, ok := object.(*v1alpha1.OBTenant); ok {
		debugWrapper(processOBTenant, tenant, "OBTenant")
	} else if server, ok := object.(*v1alpha1.OBServer); ok {
		debugWrapper(processOBServer, server, "OBServer")
	} else if zone, ok := object.(*v1alpha1.OBZone); ok {
		debugWrapper(processOBZone, zone, "OBZone")
	} else if restore, ok := object.(*v1alpha1.OBTenantRestore); ok {
		debugWrapper(processOBTenantRestore, restore, "OBTenantRestore")
	} else if policy, ok := object.(*v1alpha1.OBTenantBackupPolicy); ok {
		debugWrapper(processOBTenantBackupPolicy, policy, "OBTenantBackupPolicy")
	} else if backup, ok := object.(*v1alpha1.OBTenantBackup); ok {
		debugWrapper(processOBTenantBackup, backup, "OBTenantBackup")
	}
}

func processOBCluster(cluster *v1alpha1.OBCluster) {
	if cluster.Spec.BackupVolume != nil && cluster.Spec.BackupVolume.Volume != nil && cluster.Spec.BackupVolume.Volume.NFS != nil {
		cluster.Spec.BackupVolume.Volume.NFS.Server = md5Hash(cluster.Spec.BackupVolume.Volume.NFS.Server)
	}
}

func processOBServer(server *v1alpha1.OBServer) {
	server.Status.PodIp = md5Hash(server.Status.PodIp)
	server.Status.NodeIp = md5Hash(server.Status.NodeIp)
}

func processOBTenant(tenant *v1alpha1.OBTenant) {
	for i := range tenant.Status.Pools {
		for j := range tenant.Status.Pools[i].Units {
			tenant.Status.Pools[i].Units[j].ServerIP = md5Hash(tenant.Status.Pools[i].Units[j].ServerIP)
			if tenant.Status.Pools[i].Units[j].Migrate.ServerIP != "" {
				tenant.Status.Pools[i].Units[j].Migrate.ServerIP = md5Hash(tenant.Status.Pools[i].Units[j].Migrate.ServerIP)
			}
		}
	}
}

func processOBZone(zone *v1alpha1.OBZone) {
	for i := range zone.Status.OBServerStatus {
		zone.Status.OBServerStatus[i].Server = md5Hash(zone.Status.OBServerStatus[i].Server)
	}
	if zone.Spec.BackupVolume != nil && zone.Spec.BackupVolume.Volume != nil && zone.Spec.BackupVolume.Volume.NFS != nil {
		zone.Spec.BackupVolume.Volume.NFS.Server = md5Hash(zone.Spec.BackupVolume.Volume.NFS.Server)
	}
}

func processOBTenantRestore(restore *v1alpha1.OBTenantRestore) {
	restore.SetAnnotations(nil)
	restore.Status.RestoreProgress = nil
}

func processOBTenantBackup(backup *v1alpha1.OBTenantBackup) {
	backup.SetAnnotations(nil)
	backup.Status.ArchiveLogJob = nil
	backup.Status.BackupJob = nil
	backup.Status.DataCleanJob = nil
}

func processOBTenantBackupPolicy(policy *v1alpha1.OBTenantBackupPolicy) {
	policy.SetAnnotations(nil)
	policy.Status.LatestArchiveLogJob = nil
	policy.Status.LatestBackupCleanJob = nil
	policy.Status.LatestFullBackupJob = nil
	policy.Status.LatestIncrementalJob = nil
}

func debugWrapper[T runtime.Object](processor func(T), object T, objectType string) {
	getLogger().Printf("[%s Before] %+v\n", objectType, object)
	processor(object)
	getLogger().Printf("[%s After] %+v\n", objectType, object)
}
