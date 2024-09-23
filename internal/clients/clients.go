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

package clients

import (
	"github.com/oceanbase/ob-operator/api/k8sv1alpha1"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients/schema"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

var (
	ClusterClient          = client.NewDynamicResourceClient[*v1alpha1.OBCluster](schema.OBClusterGVR, schema.OBClusterKind)
	ZoneClient             = client.NewDynamicResourceClient[*v1alpha1.OBZone](schema.OBZoneGVR, schema.OBZoneKind)
	ServerClient           = client.NewDynamicResourceClient[*v1alpha1.OBServer](schema.OBServerGVR, schema.OBServerKind)
	TenantClient           = client.NewDynamicResourceClient[*v1alpha1.OBTenant](schema.OBTenantGVR, schema.OBTenantKind)
	BackupJobClient        = client.NewDynamicResourceClient[*v1alpha1.OBTenantBackup](schema.OBTenantBackupGVR, schema.OBTenantBackupKind)
	OperationClient        = client.NewDynamicResourceClient[*v1alpha1.OBTenantOperation](schema.OBTenantOperationGVR, schema.OBTenantOperationKind)
	BackupPolicyClient     = client.NewDynamicResourceClient[*v1alpha1.OBTenantBackupPolicy](schema.OBTenantBackupPolicyGVR, schema.OBTenantBackupPolicyKind)
	RescueClient           = client.NewDynamicResourceClient[*v1alpha1.OBResourceRescue](schema.OBResourceRescueGVR, schema.OBResourceRescueKind)
	RestoreJobClient       = client.NewDynamicResourceClient[*v1alpha1.OBTenantRestore](schema.OBTenantRestoreGVR, schema.OBTenantRestoreKind)
	ClusterOperationClient = client.NewDynamicResourceClient[*v1alpha1.OBClusterOperation](schema.OBClusterOperationGVR, schema.OBClusterOperationKind)

	K8sClusterCredentialClient = client.NewDynamicResourceClient[*k8sv1alpha1.K8sClusterCredential](schema.K8sClusterCredentialGVR, schema.K8sClusterCredentialKind)
)
