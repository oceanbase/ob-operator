package param

type CreateOBTenantParam struct {
	Name             string `json:"name" binding:"required"`
	ClusterName      string `json:"obcluster" binding:"required"`
	TenantName       string `json:"tenantName" binding:"required"`
	UnitNumber       int    `json:"unitNum" binding:"required"`
	RootPassword     string `json:"rootPassword" binding:"required"`
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`
	Charset          string `json:"charset,omitempty"`

	UnitConfig *UnitConfig        `json:"unitConfig" binding:"required"`
	Pools      []ResourcePoolSpec `json:"pools" binding:"required"`

	// Enum: Primary, Standby
	TenantRole TenantRole        `json:"tenantRole,omitempty"`
	Source     *TenantSourceSpec `json:"source,omitempty"`
}

type UpdateOBTenantParam CreateOBTenantParam

type ResourcePoolSpec struct {
	Zone     string `json:"zone" binding:"required"`
	Priority int    `json:"priority,omitempty"`
	// Enum: Readonly, Full
	Type string `json:"type,omitempty"`
}

type TenantSourceSpec struct {
	Tenant  *string            `json:"tenant,omitempty"`
	Restore *RestoreSourceSpec `json:"restore,omitempty"`
}

type RestoreSourceSpec struct {
	// Enum: OSS, NFS
	Type          BackupDestType `json:"type"`
	ArchiveSource string         `json:"archiveSource"`
	BakDataSource string         `json:"bakDataSource"`
	OSSAccessID   string         `json:"ossAccessId,omitempty"`
	OSSAccessKey  string         `json:"ossAccessKey,omitempty"`

	BakEncryptionPassword string              `json:"bakEncryptionPassword,omitempty"`
	Until                 *RestoreUntilConfig `json:"until,omitempty"`
}

type UnitConfig struct {
	CPUCount    string `json:"cpuCount" binding:"required"`
	MemorySize  string `json:"memorySize" binding:"required"`
	MaxIops     int    `json:"maxIops,omitempty"`
	MinIops     int    `json:"minIops,omitempty"`
	IopsWeight  int    `json:"iopsWeight,omitempty"`
	LogDiskSize string `json:"logDiskSize,omitempty"`
}

type RestoreUntilConfig struct {
	Timestamp *string `json:"timestamp,omitempty"`
	Unlimited bool    `json:"unlimited,omitempty"`
}

type ModifyUnitNumber struct {
	UnitNumber int `json:"unitNum" binding:"required"`
}

type ChangeRootPassword struct {
	RootPassword string `json:"rootPassword" binding:"required"`
}

type ReplayStandbyLog RestoreUntilConfig

type ChangeTenantRole struct {
	// Enum: Primary, Standby
	TenantRole TenantRole `json:"tenantRole" binding:"required"`
	Switchover bool       `json:"switchover,omitempty"`
}
