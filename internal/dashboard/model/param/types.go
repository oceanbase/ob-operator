package param

type TenantRole string
type BackupDestType string

type BackupDestination struct {
	Type            BackupDestType `json:"type,omitempty"`
	Path            string         `json:"path,omitempty"`
	OSSAccessSecret string         `json:"ossAccessSecret,omitempty"`
}

type NamespacedName struct {
	Namespace string `json:"namespace" uri:"namespace" binding:"required"`
	Name      string `json:"name" uri:"name" binding:"required"`
}
