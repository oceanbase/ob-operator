package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBServerGroup    = "oceanbase.oceanbase.com"
	OBServerVersion  = "v1alpha1"
	OBServerKind     = "OBServer"
	OBServerResource = "observers"
)

var (
	OBServerRes = schema.GroupVersionResource{
		Group:    OBServerGroup,
		Version:  OBServerVersion,
		Resource: OBServerResource,
	}
)
