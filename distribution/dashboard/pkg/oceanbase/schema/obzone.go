package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBZoneGroup    = "oceanbase.oceanbase.com"
	OBZoneVersion  = "v1alpha1"
	OBZoneKind     = "OBZone"
	OBZoneResource = "obzones"
)

var (
	OBZoneRes = schema.GroupVersionResource{
		Group:    OBZoneGroup,
		Version:  OBZoneVersion,
		Resource: OBZoneResource,
	}
)
