package constant

const (
	ConfigVolumeName          = "config"
	ConfigMountPath           = "/etc/config"
	TTLSecondsAfterFinished   = 7 * 24 * 60 * 60
	ClusterRoleName           = "oceanbase-dashboard-cluster-role"
	ServiceAccountNameFmt     = "ob-ins-%s"
	ClusterRoleBindingNameFmt = "ob-ins-%s-%s"
)
