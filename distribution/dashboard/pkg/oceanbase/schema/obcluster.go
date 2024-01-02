package schema

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	OBClusterGroup    = "oceanbase.oceanbase.com"
	OBClusterVersion  = "v1alpha1"
	OBClusterKind     = "OBCluster"
	OBClusterResource = "obclusters"
)

var (
	OBClusterRes = schema.GroupVersionResource{
		Group:    OBClusterGroup,
		Version:  OBClusterVersion,
		Resource: OBClusterResource,
	}

	OBClusterResKind = schema.GroupVersionKind{
		Group:   OBClusterGroup,
		Version: OBClusterVersion,
		Kind:    OBClusterKind,
	}
)
