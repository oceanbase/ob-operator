/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
//go:generate task_register $GOFILE

package oblogservicecluster

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicezone"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskMap = builder.NewTaskHub[*OBLogServiceClusterManager]()

func CreateZones(m *OBLogServiceClusterManager) tasktypes.TaskError {
	m.Logger.Info("Creating log service zones")
	existingZones, err := m.listZones()
	if err != nil {
		return errors.Wrap(err, "list existing zones")
	}

	existingZoneMap := make(map[string]bool)
	for _, zone := range existingZones.Items {
		existingZoneMap[zone.Spec.Topology.Zone] = true
	}

	blockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		APIVersion:         m.Resource.APIVersion,
		Kind:               m.Resource.Kind,
		Name:               m.Resource.Name,
		UID:                m.Resource.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}

	for _, topo := range m.Resource.Spec.Topology {
		if existingZoneMap[topo.Zone] {
			continue
		}
		zoneName := fmt.Sprintf("%s-%s", m.Resource.Name, topo.Zone)
		zone := &v1alpha1.OBLogServiceZone{
			ObjectMeta: metav1.ObjectMeta{
				Name:            zoneName,
				Namespace:       m.Resource.Namespace,
				OwnerReferences: []metav1.OwnerReference{ownerRef},
				Labels: map[string]string{
					oceanbaseconst.LabelRefOBLogServiceCluster: m.Resource.Name,
				},
			},
			Spec: v1alpha1.OBLogServiceZoneSpec{
				ClusterName:    m.Resource.Name,
				ClusterId:      m.Resource.Spec.ClusterId,
				Image:          m.Resource.Spec.Image,
				Resource:       m.Resource.Spec.Resource,
				Topology:       topo,
				ObjectStoreURL: m.Resource.Spec.ObjectStoreURL,
				Storage:        m.Resource.Spec.Storage,
				Parameters:     m.Resource.Spec.Parameters,
				ServiceAccount: m.Resource.Spec.ServiceAccount,
			},
		}
		if err := m.Client.Create(m.Ctx, zone); err != nil {
			if !kubeerrors.IsAlreadyExists(err) {
				return errors.Wrapf(err, "create zone %s", zoneName)
			}
		}
		m.Logger.Info("Created log service zone", "name", zoneName)
	}
	return nil
}

func WaitZonesBootstrapReady(m *OBLogServiceClusterManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for all zones to be bootstrap ready")
	for i := 0; i < 600; i++ {
		zoneList, err := m.listZones()
		if err != nil {
			return errors.Wrap(err, "list zones")
		}
		if len(zoneList.Items) < len(m.Resource.Spec.Topology) {
			time.Sleep(time.Second)
			continue
		}
		allReady := true
		for _, zone := range zoneList.Items {
			// Accept both "bootstrap ready" and "running" as zones may have already
			// progressed past bootstrap ready before the cluster task checks
			if zone.Status.Status != zonestatus.BootstrapReady && zone.Status.Status != zonestatus.Running {
				allReady = false
				break
			}
		}
		if allReady {
			m.Logger.Info("All zones are bootstrap ready")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for zones to be bootstrap ready")
}

func WaitZonesRunning(m *OBLogServiceClusterManager) tasktypes.TaskError {
	m.Logger.Info("Waiting for all zones to be running")
	for i := 0; i < 600; i++ {
		zoneList, err := m.listZones()
		if err != nil {
			return errors.Wrap(err, "list zones")
		}
		allRunning := true
		for _, zone := range zoneList.Items {
			if zone.Status.Status != zonestatus.Running {
				allRunning = false
				break
			}
		}
		if allRunning && len(zoneList.Items) > 0 {
			m.Logger.Info("All zones are running")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout waiting for zones to be running")
}

func BootstrapLogService(m *OBLogServiceClusterManager) tasktypes.TaskError {
	m.Logger.Info("Bootstrapping log service cluster")

	// Collect all node addresses from all zones
	zoneList, err := m.listZones()
	if err != nil {
		return errors.Wrap(err, "list zones for bootstrap")
	}

	type nodeInfo struct {
		region  string
		zone    string
		addr    string
		rpcPort int32
	}

	var nodes []nodeInfo
	for _, zone := range zoneList.Items {
		nodeList := &v1alpha1.OBLogServiceNodeList{}
		if err := m.Client.List(m.Ctx, nodeList, client.MatchingLabels{
			oceanbaseconst.LabelRefOBLogServiceZone: zone.Name,
		}, client.InNamespace(m.Resource.Namespace)); err != nil {
			return errors.Wrapf(err, "list nodes for zone %s", zone.Name)
		}
		for _, node := range nodeList.Items {
			// Use PodIP for bootstrap because oblogservice binds directly to local_ip (PodIP).
			// Unlike observer which separates bind addr from advertise addr, oblogservice
			// requires SERVER address to match its local_ip for election to work.
			addr := node.Status.PodIP
			rpcPort := node.Spec.RpcPort
			if rpcPort == 0 {
				rpcPort = int32(oceanbaseconst.LogServiceRpcPort)
			}
			nodes = append(nodes, nodeInfo{
				region:  oceanbaseconst.LogServiceDefaultRegion,
				zone:    node.Spec.Zone,
				addr:    addr,
				rpcPort: rpcPort,
			})
		}
	}

	if len(nodes) == 0 {
		return errors.New("no nodes available for bootstrap")
	}

	// Use the first node's HTTP address as bootstrap target
	firstZone := zoneList.Items[0]
	firstNodeList := &v1alpha1.OBLogServiceNodeList{}
	if err := m.Client.List(m.Ctx, firstNodeList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBLogServiceZone: firstZone.Name,
	}, client.InNamespace(m.Resource.Namespace)); err != nil {
		return errors.Wrap(err, "list first zone nodes")
	}
	if len(firstNodeList.Items) == 0 {
		return errors.New("first zone has no nodes")
	}
	firstNode := firstNodeList.Items[0]
	httpPort := firstNode.Spec.HttpPort
	if httpPort == 0 {
		httpPort = int32(oceanbaseconst.LogServiceHttpPort)
	}
	httpAddr := fmt.Sprintf("%s:%d", firstNode.Status.PodIP, httpPort)

	// Build server args — use env vars for user-supplied values to prevent shell injection
	serverArgs := ""
	var serverEnvVars []corev1.EnvVar
	for i, n := range nodes {
		if i > 0 {
			serverArgs += ", "
		}
		regionEnv := fmt.Sprintf("NODE_REGION_%d", i)
		zoneEnv := fmt.Sprintf("NODE_ZONE_%d", i)
		serverArgs += fmt.Sprintf("REGION ${%s} AZ ${%s} SERVER %s:%d",
			regionEnv, zoneEnv, n.addr, n.rpcPort)
		serverEnvVars = append(serverEnvVars,
			corev1.EnvVar{Name: regionEnv, Value: n.region},
			corev1.EnvVar{Name: zoneEnv, Value: n.zone},
		)
	}

	secretMountPath := "/etc/oss-credentials"
	bootstrapCmd := fmt.Sprintf(
		`ACCESS_ID=$(cat %s/access_id) && ACCESS_KEY=$(cat %s/access_key) && /home/admin/oblogservice/bin/ls_ctrl --host %s bootstrap --object-store-url "${BUCKET_URL}&access_id=${ACCESS_ID}&access_key=${ACCESS_KEY}" %s`,
		secretMountPath, secretMountPath,
		httpAddr, serverArgs,
	)

	serverEnvVars = append(serverEnvVars, corev1.EnvVar{
		Name:  "BUCKET_URL",
		Value: m.Resource.Spec.ObjectStoreURL.BucketURL,
	})

	m.Logger.Info("Execute log service bootstrap",
		"host", httpAddr,
		"bucketURL", m.Resource.Spec.ObjectStoreURL.BucketURL,
		"serverArgs", serverArgs)

	secretVolumeName := "oss-credentials"
	volumeConfig := resourceutils.JobContainerVolumes{
		VolumeMounts: []corev1.VolumeMount{{
			Name:      secretVolumeName,
			MountPath: secretMountPath,
			ReadOnly:  true,
		}},
		Volumes: []corev1.Volume{{
			Name: secretVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: m.Resource.Spec.ObjectStoreURL.SecretRef.Name,
				},
			},
		}},
		Env: serverEnvVars,
	}

	jobName := fmt.Sprintf("%s-ls-bootstrap", m.Resource.Name)
	_, exitCode, jobErr := resourceutils.RunJob(
		m.Ctx, m.Client, m.Logger, m.Resource.Namespace,
		jobName,
		m.Resource.Spec.Image,
		nil,
		bootstrapCmd,
		volumeConfig,
	)
	if jobErr != nil {
		return errors.Wrapf(jobErr, "run log service bootstrap job, exitCode=%d", exitCode)
	}

	m.Logger.Info("Log service bootstrap succeeded")
	m.Recorder.Event(m.Resource, "Bootstrap", "", "Log service bootstrap succeeded")

	// Mark all bootstrap nodes as registered
	for _, zone := range zoneList.Items {
		nodeList := &v1alpha1.OBLogServiceNodeList{}
		if err := m.Client.List(m.Ctx, nodeList, client.MatchingLabels{
			oceanbaseconst.LabelRefOBLogServiceZone: zone.Name,
		}, client.InNamespace(m.Resource.Namespace)); err != nil {
			return errors.Wrapf(err, "list nodes for zone %s annotation", zone.Name)
		}
		for i := range nodeList.Items {
			node := &nodeList.Items[i]
			patch := client.MergeFrom(node.DeepCopy())
			if node.Annotations == nil {
				node.Annotations = make(map[string]string)
			}
			node.Annotations[oceanbaseconst.AnnotationsLogServiceNodeRegistered] = "true"
			if err := m.Client.Patch(m.Ctx, node, patch); err != nil {
				return errors.Wrapf(err, "annotate node %s as registered", node.Name)
			}
		}
	}
	return nil
}

func ModifyZoneReplica(m *OBLogServiceClusterManager) tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		zoneList, err := m.listZones()
		if err != nil {
			return errors.Wrap(err, "list zones")
		}
		for _, topo := range m.Resource.Spec.Topology {
			for i := range zoneList.Items {
				zone := &zoneList.Items[i]
				if topo.Zone == zone.Spec.Topology.Zone && topo.Replica != zone.Spec.Topology.Replica {
					m.Logger.Info("Modify zone replica", "zone", topo.Zone, "from", zone.Spec.Topology.Replica, "to", topo.Replica)
					zone.Spec.Topology.Replica = topo.Replica
					if err := m.Client.Update(m.Ctx, zone); err != nil {
						return errors.Wrapf(err, "modify zone %s replica", topo.Zone)
					}
				}
			}
		}
		return nil
	})
}
