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

package obzone

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func (m *OBZoneManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
}

func (m *OBZoneManager) generateServerName() string {
	return fmt.Sprintf("%s-%d-%s-%s", m.OBZone.Spec.ClusterName, m.OBZone.Spec.ClusterId, m.OBZone.Spec.Topology.Zone, rand.String(6))
}

func (m *OBZoneManager) checkIfStorageSizeExpand(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 ||
		observer.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 ||
		observer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(m.OBZone.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0
}

func (m *OBZoneManager) checkIfStorageClassChanged(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Storage.DataStorage.StorageClass != m.OBZone.Spec.OBServerTemplate.Storage.DataStorage.StorageClass ||
		observer.Spec.OBServerTemplate.Storage.LogStorage.StorageClass != m.OBZone.Spec.OBServerTemplate.Storage.LogStorage.StorageClass ||
		observer.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass != m.OBZone.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass
}

func (m *OBZoneManager) checkIfCalcResourceChange(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		observer.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Memory) != 0
}

func (m *OBZoneManager) checkIfBackupVolumeMutated(observer *v1alpha1.OBServer) bool {
	return (observer.Spec.BackupVolume == nil) != (m.OBZone.Spec.BackupVolume == nil)
}

func (m *OBZoneManager) checkIfMonitorMutated(observer *v1alpha1.OBServer) bool {
	return (observer.Spec.MonitorTemplate == nil) != (m.OBZone.Spec.MonitorTemplate == nil)
}

func (m *OBZoneManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obzone, err := m.getOBZone()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		obzone.Status = *m.OBZone.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, obzone)
	})
}

func (m *OBZoneManager) listOBServers() (*v1alpha1.OBServerList, error) {
	// this label always exists
	observerList := &v1alpha1.OBServerList{}
	err := m.Client.List(m.Ctx, observerList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBZone: m.OBZone.Name,
	}, client.InNamespace(m.OBZone.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "get observers")
	}
	return observerList, err
}

func (m *OBZoneManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBZone.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBZoneManager) getOBZone() (*v1alpha1.OBZone, error) {
	// this label always exists
	obzone := &v1alpha1.OBZone{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBZone.Name), obzone)
	if err != nil {
		return nil, errors.Wrap(err, "get obzone")
	}
	return obzone, nil
}

func (m *OBZoneManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBZone.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBZoneManager) generateWaitOBServerStatusFunc(status string, timeoutSeconds int) tasktypes.TaskFunc {
	f := func() tasktypes.TaskError {
		for i := 1; i < timeoutSeconds; i++ {
			obzone, err := m.getOBZone()
			if err != nil {
				return errors.Wrap(err, "get obzoen failed")
			}
			allMatched := true
			for _, observerStatus := range obzone.Status.OBServerStatus {
				if observerStatus.Status != status && observerStatus.Status != serverstatus.Unrecoverable {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Server status still not matched", "server", observerStatus.Server, "status", status)
					allMatched = false
					break
				}
			}
			if allMatched {
				return nil
			}
			time.Sleep(time.Second)
		}
		return errors.New("all server still not bootstrap ready when timeout")
	}
	return f
}

func (m *OBZoneManager) createOneOBServer(serverName string) (*v1alpha1.OBServer, error) {
	blockOwnerDeletion := true
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion:         m.OBZone.APIVersion,
		Kind:               m.OBZone.Kind,
		Name:               m.OBZone.Name,
		UID:                m.OBZone.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	independentVolumeAnnoVal, independentVolumeAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsIndependentPVCLifecycle)
	singlePVCAnnoVal, singlePVCAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsSinglePVC)
	modeAnnoVal, modeAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsMode)
	migrateAnnoVal, migrateAnnoExist := resourceutils.GetAnnotationField(m.OBZone, oceanbaseconst.AnnotationsSourceClusterAddress)
	finalizerName := oceanbaseconst.FinalizerDeleteOBServer
	finalizers := []string{finalizerName}
	labels := make(map[string]string)
	cluster, _ := m.OBZone.Labels[oceanbaseconst.LabelRefOBCluster]
	labels[oceanbaseconst.LabelRefUID] = string(m.OBZone.GetUID())
	labels[oceanbaseconst.LabelRefOBZone] = m.OBZone.Name
	labels[oceanbaseconst.LabelRefOBCluster] = cluster
	observer := &v1alpha1.OBServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:            serverName,
			Namespace:       m.OBZone.Namespace,
			OwnerReferences: ownerReferenceList,
			Finalizers:      finalizers,
			Labels:          labels,
		},
		Spec: v1alpha1.OBServerSpec{
			ClusterName:          m.OBZone.Spec.ClusterName,
			ClusterId:            m.OBZone.Spec.ClusterId,
			Zone:                 m.OBZone.Spec.Topology.Zone,
			NodeSelector:         m.OBZone.Spec.Topology.NodeSelector,
			Affinity:             m.OBZone.Spec.Topology.Affinity,
			Tolerations:          m.OBZone.Spec.Topology.Tolerations,
			OBServerTemplate:     m.OBZone.Spec.OBServerTemplate,
			MonitorTemplate:      m.OBZone.Spec.MonitorTemplate,
			BackupVolume:         m.OBZone.Spec.BackupVolume,
			ServiceAccount:       m.OBZone.Spec.ServiceAccount,
			K8sClusterCredential: m.OBZone.Spec.Topology.K8sClusterCredential,
		},
	}
	zoneTopo := m.OBZone.Spec.Topology
	if zoneTopo.OBServerTemplate != nil {
		observer.Spec.OBServerTemplate = zoneTopo.OBServerTemplate
	}
	if zoneTopo.MonitorTemplate != nil {
		observer.Spec.MonitorTemplate = zoneTopo.MonitorTemplate
	}
	observer.ObjectMeta.Annotations = make(map[string]string)
	if independentVolumeAnnoExist {
		observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsIndependentPVCLifecycle] = independentVolumeAnnoVal
	}
	if singlePVCAnnoExist {
		observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSinglePVC] = singlePVCAnnoVal
	}
	if modeAnnoExist {
		observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsMode] = modeAnnoVal
	}
	if migrateAnnoExist {
		observer.ObjectMeta.Annotations[oceanbaseconst.AnnotationsSourceClusterAddress] = migrateAnnoVal
	}
	m.Logger.Info("Create observer", "server", serverName)
	err := m.Client.Create(m.Ctx, observer)
	if err != nil {
		m.Logger.Error(err, "Create observer failed", "server", serverName)
		return nil, errors.Wrap(err, "create observer")
	}
	m.Recorder.Event(m.OBZone, "CreateObServer", "CreateObserver", fmt.Sprintf("Create observer %s", serverName))
	return observer, nil
}
