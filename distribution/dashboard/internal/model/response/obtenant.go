package response

// @Description Brief information about OBTenant
type OBTenantBrief struct {
	Name        string            `json:"name"`        // Name of the resource
	Namespace   string            `json:"namespace"`   // Namespace of the resource
	TenantName  string            `json:"tenantName"`  // Name of the tenant in the database
	ClusterName string            `json:"clusterName"` // Name of the cluster belonging to
	TenantRole  string            `json:"tenantRole"`  // Enum: Primary, Standby
	UnitNumber  int               `json:"unitNumber"`  // Number of units in every zone
	Topology    []OBTenantReplica `json:"topology"`    // Topology of the tenant
	Status      string            `json:"status"`      // Status of the tenant
	CreateTime  string            `json:"createTime"`  // Creation time of the tenant
}

type OBTenantDetail struct {
	OBTenantBrief       `json:",inline"`
	RootCredential      string `json:"rootCredential"`
	StandbyROCredentail string `json:"standbyROCredentail"`

	PrimaryTenant string         `json:"primaryTenant"`
	RestoreSource *RestoreSource `json:"restoreSource,omitempty"`
}

type OBTenantReplica struct {
	Zone     string `json:"zone"`
	Priority int    `json:"priority"`
	// Enum: Readonly, Full
	Type        string `json:"type"`
	MaxCPU      string `json:"maxCPU"`
	MemorySize  string `json:"memorySize"`
	MinCPU      string `json:"minCPU,omitempty"`
	MaxIops     int    `json:"maxIops,omitempty"`
	MinIops     int    `json:"minIops,omitempty"`
	IopsWeight  int    `json:"iopsWeight,omitempty"`
	LogDiskSize string `json:"logDiskSize,omitempty"`
}

type RestoreSource struct {
	// Enum: OSS, NFS
	Type                string `json:"type"`
	ArchiveSource       string `json:"archiveSource"`
	BakDataSource       string `json:"bakDataSource"`
	OssAccessSecret     string `json:"ossAccessSecret,omitempty"`
	BakEncryptionSecret string `json:"bakEncryptionSecret,omitempty"`
	Until               string `json:"until,omitempty"`
}
