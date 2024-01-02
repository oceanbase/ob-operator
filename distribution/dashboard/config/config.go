package config

type mode string

const (
	InCluster      mode = "incluster"
	OutsideCluster mode = "outsidecluster"
)

// for test outside K8s cluster
var RunMode mode = InCluster
