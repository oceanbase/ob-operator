package tenantstatus

const (
	Creating = "creating"
	Running  = "running"
	MaintainingWhiteList="maintaining whitelist"
	MaintainingCharset="maintaining charset"
	MaintainingUnitNum="maintaining unit num"
	MaintainingPrimaryZone ="maintaining primary zone"
	MaintainingLocality    ="maintaining locality"
	AddingPool             ="adding pool"
	DeletingPool="deleting pool"
	MaintainingUnitConfig="maintaining unit config"
	Deleting = "deleting"
	FinalizerFinished = "finalizer finished"
	PausingReconcile  ="pausing reconcile"
)