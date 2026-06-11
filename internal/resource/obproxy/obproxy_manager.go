/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OR ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obproxy

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	proxystatus "github.com/oceanbase/ob-operator/internal/const/status/obproxy"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var _ opresource.ResourceManager = &OBProxyManager{}

type OBProxyManager struct {
	Ctx      context.Context
	OBProxy  *v1alpha1.OBProxy
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBProxyManager) GetMeta() metav1.Object {
	return m.OBProxy.GetObjectMeta()
}

func (m *OBProxyManager) GetStatus() string {
	return m.OBProxy.Status.Status
}

func (m *OBProxyManager) InitStatus() {
	m.Logger.Info("Newly created obproxy, init status")
	m.Recorder.Event(m.OBProxy, "Init", "", "Newly created obproxy, init status")
	status := v1alpha1.OBProxyStatus{
		Status:   proxystatus.New,
		Image:    m.OBProxy.Spec.Image,
		Replicas: m.OBProxy.Spec.Replicas,
	}
	m.OBProxy.Status = status
}

func (m *OBProxyManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.OBProxy.Status.OperationContext = c
}

func (m *OBProxyManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBProxy.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow from obproxy status")
		return tasktypes.NewTaskFlow(m.OBProxy.Status.OperationContext), nil
	}

	// return task flow depends on status
	var taskFlow *tasktypes.TaskFlow
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Create task flow according to obproxy status")

	switch m.OBProxy.Status.Status {
	case proxystatus.New:
		obcluster, err := m.getOBCluster()
		if err != nil {
			m.Logger.Info("OBCluster not found, waiting", "cluster", m.OBProxy.Spec.OBCluster.Name)
			return nil, nil
		}
		if obcluster.Status.Status != clusterstatus.Running {
			m.Logger.Info("OBCluster not ready, waiting", "cluster", obcluster.Name, "status", obcluster.Status.Status)
			return nil, nil
		}
		taskFlow = genCreateOBProxyFlow(m)
	case proxystatus.Updating:
		taskFlow = genUpdateOBProxyFlow(m)
	case proxystatus.Scaling:
		taskFlow = genScaleOBProxyFlow(m)
	case proxystatus.Deleting:
		taskFlow = genDeleteOBProxyFlow(m)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("No need to run anything for obproxy", "obproxy", m.OBProxy.Name)
		return nil, nil
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = proxystatus.Running
		}
	}

	return taskFlow, nil
}

func (m *OBProxyManager) CheckAndUpdateFinalizers() error {
	if m.OBProxy.Status.Status == proxystatus.FinalizerFinished {
		m.OBProxy.ObjectMeta.Finalizers = make([]string, 0)
		return m.Client.Update(m.Ctx, m.OBProxy)
	}
	return nil
}

func (m *OBProxyManager) UpdateStatus() error {
	deployment, err := m.getOBProxyDeployment()
	if err != nil {
		m.Logger.Error(err, "get obproxy deployment error")
		return errors.Wrap(err, "get obproxy deployment")
	}
	if deployment != nil {
		m.OBProxy.Status.Image = deployment.Spec.Template.Spec.Containers[0].Image
		if deployment.Spec.Replicas != nil {
			m.OBProxy.Status.Replicas = *deployment.Spec.Replicas
		}
		m.OBProxy.Status.ReadyReplicas = deployment.Status.ReadyReplicas
		// Read RS_LIST env from the running Deployment (observed, not desired).
		for _, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name == "obproxy" {
				for _, env := range c.Env {
					if env.Name == "RS_LIST" {
						m.OBProxy.Status.RSList = env.Value
						break
					}
				}
				break
			}
		}
	}

	if svc, svcErr := m.getOBProxyService(); svcErr == nil && svc != nil {
		m.OBProxy.Status.ServiceIP = svc.Spec.ClusterIP
	}

	obcluster, clusterErr := m.getOBCluster()
	if clusterErr != nil {
		m.setCondition("OBClusterAvailable", false, "NotFound", clusterErr.Error())
		m.setCondition("OBClusterReady", false, "Unknown", "OBCluster not reachable")
		m.Logger.V(1).Info("cannot get obcluster", "error", clusterErr)
	} else {
		m.setCondition("OBClusterAvailable", true, "Found", "")
		if obcluster.Status.Status == clusterstatus.Running && obcluster.Status.OperationContext == nil {
			m.setCondition("OBClusterReady", true, "Running", "")
		} else {
			m.setCondition("OBClusterReady", false, "Transitioning", obcluster.Status.Status)
		}
	}

	if deployment != nil && m.OBProxy.Status.Status == proxystatus.Running {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == "obproxy" && container.Image != m.OBProxy.Spec.Image {
				m.Logger.Info("OBProxy image changed, need update")
				m.OBProxy.Status.Status = proxystatus.Updating
				break
			}
		}

		if m.OBProxy.Status.Status == proxystatus.Running {
			if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas != m.OBProxy.Spec.Replicas {
				m.Logger.Info("OBProxy replicas changed, need scale")
				m.OBProxy.Status.Status = proxystatus.Scaling
			}
		}

		if m.OBProxy.Status.Status == proxystatus.Running && m.OBProxy.Spec.Resource != nil {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				if container.Name == "obproxy" {
					if !m.isResourceEqual(container.Resources, *m.OBProxy.Spec.Resource) {
						m.Logger.Info("OBProxy resource changed, need update")
						m.OBProxy.Status.Status = proxystatus.Updating
						break
					}
				}
			}
		}

		if m.OBProxy.Status.Status == proxystatus.Running {
			podSpec := deployment.Spec.Template.Spec
			if !reflect.DeepEqual(podSpec.NodeSelector, m.OBProxy.Spec.NodeSelector) ||
				!reflect.DeepEqual(podSpec.Affinity, m.OBProxy.Spec.Affinity) ||
				!reflect.DeepEqual(podSpec.Tolerations, m.OBProxy.Spec.Tolerations) {
				m.Logger.Info("OBProxy scheduling constraints changed, need update")
				m.OBProxy.Status.Status = proxystatus.Updating
			}
		}

		if m.OBProxy.Status.Status == proxystatus.Running {
			if clusterErr != nil {
				m.setCondition("RSListAvailable", false, "OBClusterUnavailable", "cannot get obcluster")
				m.Logger.V(1).Info("skip RS_LIST drift check, cannot get obcluster", "error", clusterErr)
			} else if obcluster.Status.Status != clusterstatus.Running || obcluster.Status.OperationContext != nil {
				m.setCondition("RSListAvailable", false, "OBClusterNotStable", obcluster.Status.Status)
				m.Logger.V(1).Info("skip RS_LIST drift check, obcluster is not stable",
					"clusterStatus", obcluster.Status.Status,
					"hasOperationContext", obcluster.Status.OperationContext != nil)
			} else {
				desiredRS, rsSrc, rsErr := m.getRootServiceList()
				if rsErr != nil {
					m.setCondition("RSListAvailable", false, "ResolveFailed", rsErr.Error())
					m.Logger.V(1).Info("skip RS_LIST drift check", "error", rsErr)
				} else {
					m.setCondition("RSListAvailable", true, "Resolved", "")
					m.OBProxy.Status.RSListSource = rsSrc
					currentRS := m.OBProxy.Status.RSList
					if desiredRS != currentRS {
						m.Logger.Info("RS_LIST drift detected, OBProxy will restart to update RS_LIST",
							"obproxy", m.OBProxy.Name,
							"namespace", m.OBProxy.Namespace,
							"desiredRS", desiredRS,
							"currentRS", currentRS)
						m.Recorder.Event(m.OBProxy, "RSListDrift", "Update",
							fmt.Sprintf("RS_LIST drift detected: %s -> %s", currentRS, desiredRS))
						m.OBProxy.Status.Status = proxystatus.Updating
					}
				}
			}
		}

		if m.OBProxy.Status.Status == proxystatus.Running {
			cm, cmErr := m.getOBProxyConfigMap()
			if cmErr == nil && cm != nil && !m.isParametersEqual(cm.Data) {
				m.Logger.Info("OBProxy parameters changed, need update")
				m.OBProxy.Status.Status = proxystatus.Updating
			}
		}
	}

	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Update obproxy status", "status", m.OBProxy.Status)
	err = m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update obproxy status")
	}
	return err
}

func (m *OBProxyManager) ClearTaskInfo() {
	m.OBProxy.Status.Status = proxystatus.Running
	m.OBProxy.Status.OperationContext = nil
}

func (m *OBProxyManager) FinishTask() {
	m.OBProxy.Status.Status = m.OBProxy.Status.OperationContext.TargetStatus
	m.OBProxy.Status.OperationContext = nil
}

func (m *OBProxyManager) HandleFailure() {
	operationContext := m.OBProxy.Status.OperationContext
	failureRule := operationContext.OnFailure
	switch failureRule.Strategy {
	case strategy.StartOver:
		if m.OBProxy.Status.Status != failureRule.NextTryStatus {
			m.OBProxy.Status.Status = failureRule.NextTryStatus
			m.OBProxy.Status.OperationContext = nil
		} else {
			m.OBProxy.Status.OperationContext.Idx = 0
			m.OBProxy.Status.OperationContext.TaskStatus = ""
			m.OBProxy.Status.OperationContext.TaskId = ""
			m.OBProxy.Status.OperationContext.Task = ""
		}
	case strategy.RetryFromCurrent:
		operationContext.TaskStatus = taskstatus.Pending
	case strategy.Pause:
	}
}

func (m *OBProxyManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	return taskMap.GetTask(name, m)
}

func (m *OBProxyManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBProxy, corev1.EventTypeWarning, "Task failed", err.Error())
}

func (m *OBProxyManager) ArchiveResource() {
	m.Logger.Info("Archive obproxy", "obproxy", m.OBProxy.Name)
	m.Recorder.Event(m.OBProxy, "Archive", "", "Archive obproxy")
	m.OBProxy.Status.Status = proxystatus.Failed
	m.OBProxy.Status.OperationContext = nil
}

// Helper functions for manager

func (m *OBProxyManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return m.Client.Status().Update(m.Ctx, m.OBProxy)
	})
}

func (m *OBProxyManager) isResourceEqual(resources corev1.ResourceRequirements, spec apitypes.ResourceSpec) bool {
	// Compare CPU
	if !spec.Cpu.IsZero() {
		if resources.Limits.Cpu().Cmp(spec.Cpu) != 0 {
			return false
		}
	}
	// Compare Memory
	if !spec.Memory.IsZero() {
		if resources.Limits.Memory().Cmp(spec.Memory) != 0 {
			return false
		}
	}
	return true
}

func (m *OBProxyManager) isParametersEqual(cmData map[string]string) bool {
	if len(cmData) != len(m.OBProxy.Spec.Parameters) {
		return false
	}
	for _, param := range m.OBProxy.Spec.Parameters {
		key := strings.ToUpper(envPrefix + param.Name)
		if cmData[key] != param.Value {
			return false
		}
	}
	return true
}

// setCondition upserts a Condition on the OBProxy status.
// It preserves LastTransitionTime when the status value has not changed.
func (m *OBProxyManager) setCondition(condType string, ok bool, reason, message string) {
	desired := metav1.ConditionFalse
	if ok {
		desired = metav1.ConditionTrue
	}
	now := metav1.Now()
	for i := range m.OBProxy.Status.Conditions {
		c := &m.OBProxy.Status.Conditions[i]
		if c.Type != condType {
			continue
		}
		if c.Status != desired {
			c.LastTransitionTime = now
		}
		c.Status = desired
		c.Reason = reason
		c.Message = message
		return
	}
	m.OBProxy.Status.Conditions = append(m.OBProxy.Status.Conditions, metav1.Condition{
		Type:               condType,
		Status:             desired,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	})
}
