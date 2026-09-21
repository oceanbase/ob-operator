package common

import (
	apitypes "github.com/oceanbase/ob-operator/api/types"
	corev1 "k8s.io/api/core/v1"
)

// ObjectReference deliberately contains no Secret data.
type ObjectReference struct {
	Name string `json:"name" binding:"required"`
}

type SharedStorageSpec struct {
	BucketURL    string          `json:"bucketURL" binding:"required"`
	SecretRef    ObjectReference `json:"secretRef" binding:"required"`
	MaxIOPS      string          `json:"maxIOPS,omitempty"`
	MaxBandwidth string          `json:"maxBandwidth,omitempty"`
}

func (s *SharedStorageSpec) ToAPI() *apitypes.SharedStorageSpec {
	if s == nil {
		return nil
	}
	return &apitypes.SharedStorageSpec{BucketURL: s.BucketURL, SecretRef: corev1.LocalObjectReference{Name: s.SecretRef.Name}, MaxIOPS: s.MaxIOPS, MaxBandwidth: s.MaxBandwidth}
}

func SharedStorageFromAPI(s *apitypes.SharedStorageSpec) *SharedStorageSpec {
	if s == nil {
		return nil
	}
	return &SharedStorageSpec{BucketURL: s.BucketURL, SecretRef: ObjectReference{Name: s.SecretRef.Name}, MaxIOPS: s.MaxIOPS, MaxBandwidth: s.MaxBandwidth}
}

func (r *ObjectReference) ToLogServiceRef() *apitypes.LogServiceReference {
	if r == nil {
		return nil
	}
	return &apitypes.LogServiceReference{Name: r.Name}
}

func LogServiceRefFromAPI(r *apitypes.LogServiceReference) *ObjectReference {
	if r == nil {
		return nil
	}
	return &ObjectReference{Name: r.Name}
}
