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

func (m *OBZoneManager) checkIfCalcResourceChange(observer *v1alpha1.OBServer) bool {
	return observer.Spec.OBServerTemplate.Resource.Cpu.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
		observer.Spec.OBServerTemplate.Resource.Memory.Cmp(m.OBZone.Spec.OBServerTemplate.Resource.Memory) != 0
}

func (m *OBZoneManager) checkIfBackupVolumeAdded(observer *v1alpha1.OBServer) bool {
	return observer.Spec.BackupVolume == nil && m.OBZone.Spec.BackupVolume != nil
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
