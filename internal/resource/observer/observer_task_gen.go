// Code generated by go generate; DO NOT EDIT.
package observer

func init() {
	taskMap.Register(tWaitOBServerReady, WaitOBServerReady)
	taskMap.Register(tAddServer, AddServer)
	taskMap.Register(tWaitOBClusterBootstrapped, WaitOBClusterBootstrapped)
	taskMap.Register(tCreateOBServerPod, CreateOBServerPod)
	taskMap.Register(tCreateOBServerPVC, CreateOBServerPVC)
	taskMap.Register(tDeleteOBServerInCluster, DeleteOBServerInCluster)
	taskMap.Register(tAnnotateOBServerPod, AnnotateOBServerPod)
	taskMap.Register(tUpgradeOBServerImage, UpgradeOBServerImage)
	taskMap.Register(tWaitOBServerPodReady, WaitOBServerPodReady)
	taskMap.Register(tWaitOBServerActiveInCluster, WaitOBServerActiveInCluster)
	taskMap.Register(tWaitOBServerDeletedInCluster, WaitOBServerDeletedInCluster)
	taskMap.Register(tDeletePod, DeletePod)
	taskMap.Register(tWaitForPodDeleted, WaitForPodDeleted)
	taskMap.Register(tExpandPVC, ExpandPVC)
	taskMap.Register(tWaitForPvcResized, WaitForPvcResized)
	taskMap.Register(tCreateOBServerSvc, CreateOBServerSvc)
	taskMap.Register(tCheckAndCreateNs, CheckAndCreateNs)
}
