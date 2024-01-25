package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBTenantBackupKind     = "OBTenantBackup"
	OBTenantBackupResource = "obtenantbackups"
)

var (
	OBTenantBackupGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: OBTenantBackupResource,
	}
	OBTenantBackupGVK = schema.GroupVersionKind{
		Group:   Group,
		Version: Version,
		Kind:    OBTenantBackupKind,
	}
)
