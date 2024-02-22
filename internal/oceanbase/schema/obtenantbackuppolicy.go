package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBTenantBackupPolicyKind     = "OBTenantBackupPolicy"
	OBTenantBackupPolicyResource = "obtenantbackuppolicies"
)

var (
	OBTenantBackupPolicyGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: OBTenantBackupPolicyResource,
	}
	OBTenantBackupPolicyGVK = schema.GroupVersionKind{
		Group:   Group,
		Version: Version,
		Kind:    OBTenantBackupPolicyKind,
	}
)
