package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBTenantGroup    = "oceanbase.oceanbase.com"
	OBTenantVersion  = "v1alpha1"
	OBTenantKind     = "OBTenant"
	OBTenantResource = "obtenants"
)

var (
	OBTenantRes = schema.GroupVersionResource{
		Group:    OBTenantGroup,
		Version:  OBTenantVersion,
		Resource: OBTenantResource,
	}
	OBTenantResKind = schema.GroupVersionKind{
		Group:   OBTenantGroup,
		Version: OBTenantVersion,
		Kind:    OBTenantKind,
	}
)
