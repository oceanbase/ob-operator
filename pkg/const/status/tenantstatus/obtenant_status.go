package tenantstatus

const (
	CreatingTenant         = "creating"
	Running                = "running"
	MaintainingWhiteList   = "maintaining whitelist"
	MaintainingCharset     = "maintaining charset"
	MaintainingUnitNum     = "maintaining unit num"
	MaintainingPrimaryZone = "maintaining primary zone"
	MaintainingLocality    = "maintaining locality"
	AddingResourcePool     = "adding resource pool"
	DeletingResourcePool   = "deleting resource pool"
	MaintainingUnitConfig  = "maintaining unit config"
	DeletingTenant         = "deleting"
	FinalizerFinished      = "finalizer finished"
	PausingReconcile       = "pausing reconcile"
)
