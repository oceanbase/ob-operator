package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBTenantOperationKind     = "OBTenantOperation"
	OBTenantOperationResource = "obtenantoperations"
)

var (
	OBTenantOperationGVR = schema.GroupVersionResource{
		Group:    Group,
		Version:  Version,
		Resource: OBTenantOperationResource,
	}
	OBTenantOperationGVK = schema.GroupVersionKind{
		Group:   Group,
		Version: Version,
		Kind:    OBTenantOperationKind,
	}
)
