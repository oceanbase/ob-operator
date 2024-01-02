package param

type CreateOBTenantParam struct {
	ClusterName      string `json:"obcluster" binding:"required"`
	TenantName       string `json:"tenantName" binding:"required"`
	UnitNumber       int    `json:"unitNum" binding:"required"`
	ForceDelete      bool   `json:"forceDelete,omitempty"`
	Charset          string `json:"charset,omitempty"`
	Collate          string `json:"collate,omitempty"`
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`

	Pools       []ResourcePoolSpec `json:"pools" binding:"required"`
	TenantRole  TenantRole         `json:"tenantRole,omitempty"`
	Source      *TenantSourceSpec  `json:"source,omitempty"`
	Credentials TenantCredentials  `json:"credentials,omitempty"`
}

type UpdateOBTenantParam CreateOBTenantParam

type ResourcePoolSpec struct {
	Zone     string `json:"zone" binding:"required"`
	Priority int    `json:"priority,omitempty"`

	Type       *LocalityType `json:"type,omitempty"`
	UnitConfig *UnitConfig   `json:"resource" binding:"required"`
}

type LocalityType struct {
	Name     string `json:"name" binding:"required"`
	Replica  int    `json:"replica" binding:"required"`
	IsActive bool   `json:"isActive" binding:"required"`
}

type TenantCredentials struct {
	Root      string `json:"root,omitempty"`
	StandbyRO string `json:"standbyRo,omitempty"`
}

type TenantSourceSpec struct {
	Tenant  *string            `json:"tenant,omitempty"`
	Restore *RestoreSourceSpec `json:"restore,omitempty"`
}

type RestoreSourceSpec struct {
	ArchiveSource       *BackupDestination `json:"archiveSource,omitempty"`
	BakDataSource       *BackupDestination `json:"bakDataSource,omitempty"`
	BakEncryptionSecret string             `json:"bakEncryptionSecret,omitempty"`

	SourceUri      string              `json:"sourceUri,omitempty"` // Deprecated
	Until          RestoreUntilConfig  `json:"until" binding:"required"`
	Description    *string             `json:"description,omitempty"`
	ReplayLogUntil *RestoreUntilConfig `json:"replayLogUntil,omitempty"`
	Cancel         bool                `json:"cancel,omitempty"`
}

type UnitConfig struct {
	MaxCPU      string `json:"maxCPU" binding:"required"`
	MemorySize  string `json:"memorySize" binding:"required"`
	MinCPU      string `json:"minCPU,omitempty"`
	MaxIops     int    `json:"maxIops,omitempty"`
	MinIops     int    `json:"minIops,omitempty"`
	IopsWeight  int    `json:"iopsWeight,omitempty"`
	LogDiskSize string `json:"logDiskSize,omitempty"`
}

type RestoreUntilConfig struct {
	Timestamp *string `json:"timestamp,omitempty"`
	Scn       *string `json:"scn,omitempty"`
	Unlimited bool    `json:"unlimited,omitempty"`
}
