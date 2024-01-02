package common

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ResourceSpec struct {
	Cpu      int64 `json:"cpu"`
	MemoryGB int64 `json:"memory"`
}

type StorageSpec struct {
	StorageClass string `json:"storageClass"`
	SizeGB       int64  `json:"size"`
}
