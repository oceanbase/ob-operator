package oceanbase

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase/schema"
)

var (
	ClusterClient      = client.NewDynamicResourceClient[*v1alpha1.OBCluster](schema.OBClusterRes, schema.OBClusterKind)
	ZoneClient         = client.NewDynamicResourceClient[*v1alpha1.OBZone](schema.OBZoneRes, schema.OBZoneKind)
	ServerClient       = client.NewDynamicResourceClient[*v1alpha1.OBServer](schema.OBServerRes, schema.OBServerKind)
	TenantClient       = client.NewDynamicResourceClient[*v1alpha1.OBTenant](schema.OBTenantRes, schema.OBTenantKind)
	BackupJobClient    = client.NewDynamicResourceClient[*v1alpha1.OBTenantBackup](schema.OBTenantBackupGVR, schema.OBTenantBackupKind)
	OperationClient    = client.NewDynamicResourceClient[*v1alpha1.OBTenantOperation](schema.OBTenantOperationGVR, schema.OBTenantOperationKind)
	BackupPolicyClient = client.NewDynamicResourceClient[*v1alpha1.OBTenantBackupPolicy](schema.OBTenantBackupPolicyGVR, schema.OBTenantBackupPolicyKind)
)
