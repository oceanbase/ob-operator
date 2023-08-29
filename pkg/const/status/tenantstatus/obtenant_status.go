package tenantstatus

const (
	CreatingTenant = "creating tenant"
	Running        = "running"
	MaintainingWhiteList="maintaining whitelist"
	MaintainingCharset="maintaining charset"
	MaintainingUnitNum="maintaining unit num"
	MaintainingPrimaryZone ="maintaining primary zone"
	MaintainingLocality    ="maintaining locality"
	AddingPool             ="adding pool"
	DeletingPool="deleting pool"
	MaintainingUnitConfig ="maintaining unit config"
	DeletingTenant        = "deleting tenant"
	FinalizerFinished     = "finalizer finished"
	PausingReconcile  ="pausing reconcile"
)