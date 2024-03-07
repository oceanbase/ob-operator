/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

// get observer from K8s api server
func (m *OBServerManager) getOBServer() (*v1alpha1.OBServer, error) {
	// this label always exists
	observer := &v1alpha1.OBServer{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), observer)
	if err != nil {
		return nil, errors.Wrap(err, "get observer")
	}
	return observer, nil
}

func (m *OBServerManager) getOBZone() (*v1alpha1.OBZone, error) {
	// this label always exists
	zoneName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBZone]
	obzone := &v1alpha1.OBZone{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(zoneName), obzone)
	if err != nil {
		return nil, errors.Wrap(err, "get obzone")
	}
	return obzone, nil
}

func (m *OBServerManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBServer.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBServerManager) getPod() (*corev1.Pod, error) {
	// this label always exists
	pod := &corev1.Pod{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), pod)
	if err != nil {
		return nil, errors.Wrap(err, "get pod")
	}
	return pod, nil
}

func (m *OBServerManager) getSvc() (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), svc)
	if err != nil {
		return nil, errors.Wrap(err, "get svc")
	}
	return svc, nil
}

func (m *OBServerManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

func (m *OBServerManager) getCurrentOBServerFromOB() (*model.OBServer, error) {
	if m.OBServer.Status.PodIp == "" {
		err := errors.New("pod ip is empty")
		m.Logger.Error(err, "unable to get observer info")
		return nil, err
	}
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	mode, modeExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if modeExist && mode == oceanbaseconst.ModeStandalone {
		observerInfo.Ip = "127.0.0.1"
	}
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrapf(err, "Get oceanbase operation manager failed")
	}
	return operationManager.GetServer(observerInfo)
}

func (m *OBServerManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		observer, err := m.getOBServer()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		observer.Status = *m.OBServer.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, observer)
	})
}

func (m *OBServerManager) setRecoveryStatus() {
	if m.OBServer.SupportStaticIP() {
		m.Logger.Info("Current server can keep static ip address or the cluster runs as standalone, recover by recreating pod")
		m.OBServer.Status.Status = serverstatus.Recover
	} else {
		m.Logger.Info("observer not recoverable, delete current observer and wait recreate")
		m.OBServer.Status.Status = serverstatus.Unrecoverable
	}
}

func (m *OBServerManager) getPVCs() (*corev1.PersistentVolumeClaimList, error) {
	pvcs := &corev1.PersistentVolumeClaimList{}
	err := m.Client.List(m.Ctx, pvcs, client.InNamespace(m.OBServer.Namespace), client.MatchingLabels{oceanbaseconst.LabelRefUID: m.OBServer.Labels[oceanbaseconst.LabelRefUID]})
	if err != nil {
		return nil, errors.Wrap(err, "list pvc")
	}
	return pvcs, nil
}

func (m *OBServerManager) checkIfStorageExpand(pvcs *corev1.PersistentVolumeClaimList) bool {
	for _, pvc := range pvcs.Items {
		switch {
		case strings.HasSuffix(pvc.Name, oceanbaseconst.DataVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 {
				return true
			}
		case strings.HasSuffix(pvc.Name, oceanbaseconst.ClogVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0 {
				return true
			}
		case strings.HasSuffix(pvc.Name, oceanbaseconst.LogVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 {
				return true
			}
		case pvc.Name == m.OBServer.Name:
			sum := resource.Quantity{}
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
			if pvc.Spec.Resources.Requests.Storage().Cmp(sum) < 0 {
				return true
			}
		}
	}
	return false
}

func (m *OBServerManager) checkIfBackupVolumeAdded(pod *corev1.Pod) bool {
	if m.OBServer.Spec.BackupVolume != nil && m.OBServer.Spec.BackupVolume.Volume != nil {
		// If the backup volume is not mounted, it means the backup volume is added
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == m.OBServer.Spec.BackupVolume.Volume.Name {
				return false
			}
		}
		return true
	}
	return false
}
