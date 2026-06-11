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
//go:generate task_register $GOFILE

package obproxy

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	observerstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskMap = builder.NewTaskHub[*OBProxyManager]()

func CopyProxyROSecret(m *OBProxyManager) tasktypes.TaskError {
	obcluster, err := m.getOBCluster()
	if err != nil {
		m.Logger.Error(err, "Failed to get obcluster")
		return errors.Wrap(err, "get obcluster")
	}

	clusterNS := m.OBProxy.Spec.OBCluster.Namespace
	if clusterNS == "" {
		clusterNS = m.OBProxy.Namespace
	}

	proxyROSecretName := obcluster.Spec.UserSecrets.ProxyRO
	if proxyROSecretName == "" {
		return errors.New("obcluster does not have proxyRO secret configured")
	}

	sourceSecret := &corev1.Secret{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: clusterNS,
		Name:      proxyROSecretName,
	}, sourceSecret)
	if err != nil {
		m.Logger.Error(err, "Failed to get proxyRO secret from obcluster namespace",
			"namespace", clusterNS, "secret", proxyROSecretName)
		return errors.Wrap(err, "get proxyRO secret from obcluster namespace")
	}

	targetSecretName := m.getProxyROSecretName()
	existingSecret := &corev1.Secret{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      targetSecretName,
	}, existingSecret)
	if err == nil {
		return nil
	}
	if !kubeerrors.IsNotFound(err) {
		return errors.Wrap(err, "check existing proxyRO secret")
	}

	newSecret := m.buildCopiedProxyROSecret(sourceSecret)
	err = m.Client.Create(m.Ctx, newSecret)
	if err != nil {
		m.Logger.Error(err, "Failed to create proxyRO secret copy")
		return errors.Wrap(err, "create proxyRO secret copy")
	}

	m.Logger.Info("Created proxyRO secret copy", "secret", targetSecretName)
	m.Recorder.Event(m.OBProxy, "CreateSecret", "", fmt.Sprintf("Created proxyRO secret %s", targetSecretName))
	return nil
}

func CreateOBProxyConfigMap(m *OBProxyManager) tasktypes.TaskError {
	cmName := m.getConfigMapName()

	existingCM := &corev1.ConfigMap{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      cmName,
	}, existingCM)
	if err == nil {
		m.Logger.Info("ConfigMap already exists", "configmap", cmName)
		return nil
	}
	if !kubeerrors.IsNotFound(err) {
		return errors.Wrap(err, "check existing configmap")
	}

	cm := m.buildConfigMap()
	err = m.Client.Create(m.Ctx, cm)
	if err != nil {
		m.Logger.Error(err, "Failed to create configmap")
		return errors.Wrap(err, "create configmap")
	}

	m.Logger.Info("Created ConfigMap", "configmap", cmName)
	m.Recorder.Event(m.OBProxy, "CreateConfigMap", "", fmt.Sprintf("Created ConfigMap %s", cmName))
	return nil
}

func CreateOBProxyService(m *OBProxyManager) tasktypes.TaskError {
	svcName := m.getServiceName()

	existingSvc := &corev1.Service{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      svcName,
	}, existingSvc)
	if err == nil {
		m.Logger.Info("Service already exists", "service", svcName)
		return nil
	}
	if !kubeerrors.IsNotFound(err) {
		return errors.Wrap(err, "check existing service")
	}

	svc := m.buildService()
	err = m.Client.Create(m.Ctx, svc)
	if err != nil {
		m.Logger.Error(err, "Failed to create service")
		return errors.Wrap(err, "create service")
	}

	m.Logger.Info("Created Service", "service", svcName)
	m.Recorder.Event(m.OBProxy, "CreateService", "", fmt.Sprintf("Created Service %s", svcName))
	return nil
}

func CreateOBProxyDeployment(m *OBProxyManager) tasktypes.TaskError {
	rsList, _, err := m.getRootServiceList()
	if err != nil {
		m.Logger.Error(err, "Failed to get rootservice list")
		return errors.Wrap(err, "get rootservice list")
	}

	svcName := m.getServiceName()
	svc := &corev1.Service{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      svcName,
	}, svc)
	if err != nil {
		m.Logger.Error(err, "Failed to get service", "service", svcName)
		return errors.Wrap(err, "get service")
	}

	proxySysSecret, err := m.getProxySysSecret()
	if err != nil {
		m.Logger.Error(err, "Failed to get proxySys secret")
		return errors.Wrap(err, "get proxySys secret")
	}

	proxyROSecretName := m.getProxyROSecretName()
	proxyROSecret := &corev1.Secret{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      proxyROSecretName,
	}, proxyROSecret)
	if err != nil {
		m.Logger.Error(err, "Failed to get proxyRO secret", "secret", proxyROSecretName)
		return errors.Wrap(err, "get proxyRO secret")
	}

	existingDeploy := &appsv1.Deployment{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      m.OBProxy.Name,
	}, existingDeploy)
	if err == nil {
		m.Logger.Info("Deployment already exists", "deployment", m.OBProxy.Name)
		return nil
	}
	if !kubeerrors.IsNotFound(err) {
		return errors.Wrap(err, "check existing deployment")
	}

	deploy := m.buildDeployment(rsList, svc, proxyROSecret, proxySysSecret)
	err = m.Client.Create(m.Ctx, deploy)
	if err != nil {
		m.Logger.Error(err, "Failed to create deployment")
		return errors.Wrap(err, "create deployment")
	}

	m.Logger.Info("Created Deployment", "deployment", m.OBProxy.Name, "rsList", rsList)
	m.Recorder.Event(m.OBProxy, "CreateDeployment", "", fmt.Sprintf("Created Deployment %s", m.OBProxy.Name))
	return nil
}

func WaitOBProxyReady(m *OBProxyManager) tasktypes.TaskError {
	timeout := 300 // 5 minutes
	for range timeout {
		deploy, err := m.getOBProxyDeployment()
		if err != nil {
			return errors.Wrap(err, "get obproxy deployment during wait")
		}
		if deploy == nil {
			return errors.New("deployment not found during wait")
		}

		if m.isOBProxyReady(deploy) {
			m.Logger.Info("OBProxy is ready",
				"obproxy", m.OBProxy.Name,
				"replicas", deploy.Status.ReadyReplicas)
			m.Recorder.Event(m.OBProxy, "Ready", "",
				fmt.Sprintf("OBProxy ready: %d/%d replicas", deploy.Status.ReadyReplicas, deploy.Status.Replicas))
			return nil
		}

		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Waiting for OBProxy to be ready",
			"ready", deploy.Status.ReadyReplicas, "desired", deploy.Status.Replicas)
		time.Sleep(time.Second)
	}

	m.Logger.Error(errors.New("timeout"), "OBProxy wait timeout", "obproxy", m.OBProxy.Name)
	return errors.New("timeout waiting for obproxy to be ready")
}

func UpdateOBProxyConfigMap(m *OBProxyManager) tasktypes.TaskError {
	m.Logger.Info("Updating OBProxy ConfigMap",
		"obproxy", m.OBProxy.Name,
		"namespace", m.OBProxy.Namespace)

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cm, err := m.getOBProxyConfigMap()
		if err != nil {
			return errors.Wrap(err, "get obproxy configmap")
		}
		if cm == nil {
			newCM := m.buildConfigMap()
			return m.Client.Create(m.Ctx, newCM)
		}

		updatedCM := m.updateConfigMapData(cm, m.OBProxy.Spec.Parameters)
		return m.Client.Update(m.Ctx, updatedCM)
	})
}

func UpdateOBProxyDeployment(m *OBProxyManager) tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deploy, err := m.getOBProxyDeployment()
		if err != nil {
			return errors.Wrap(err, "get obproxy deployment")
		}
		if deploy == nil {
			return errors.New("deployment not found")
		}

		m.Logger.Info("Updating OBProxy Deployment",
			"obproxy", m.OBProxy.Name,
			"namespace", m.OBProxy.Namespace,
			"deployment", deploy.Name,
			"replicas", m.OBProxy.Spec.Replicas,
			"image", m.OBProxy.Spec.Image)

		if deploy.Spec.Template.Spec.Containers[0].Image != m.OBProxy.Spec.Image {
			deploy.Spec.Template.Spec.Containers[0].Image = m.OBProxy.Spec.Image
		}

		if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas != m.OBProxy.Spec.Replicas {
			replicas := m.OBProxy.Spec.Replicas
			deploy.Spec.Replicas = &replicas
		}

		rsList, _, err := m.getRootServiceList()
		if err != nil {
			return errors.Wrap(err, "get rootservice list")
		}
		for i := range deploy.Spec.Template.Spec.Containers {
			container := &deploy.Spec.Template.Spec.Containers[i]
			if container.Name == "obproxy" {
				for j := range container.Env {
					if container.Env[j].Name == "RS_LIST" {
						if container.Env[j].Value != rsList {
							container.Env[j].Value = rsList
						}
						break
					}
				}

				container.Resources = m.buildResourceRequirements()

				var newMemLimited string
				if m.OBProxy.Spec.Resource != nil && !m.OBProxy.Spec.Resource.Memory.IsZero() {
					memoryMB := m.OBProxy.Spec.Resource.Memory.Value() * 95 / 100 / (1 << 20)
					newMemLimited = fmt.Sprintf("%dMB", memoryMB)
				}
				memIdx := -1
				for j := range container.Env {
					if container.Env[j].Name == "ODP_PROXY_MEM_LIMITED" {
						memIdx = j
						break
					}
				}
				if memIdx >= 0 {
					if newMemLimited == "" {
						container.Env = append(container.Env[:memIdx], container.Env[memIdx+1:]...)
					} else {
						container.Env[memIdx].Value = newMemLimited
					}
				} else if newMemLimited != "" {
					container.Env = append(container.Env, corev1.EnvVar{Name: "ODP_PROXY_MEM_LIMITED", Value: newMemLimited})
				}

				break
			}
		}

		deploy.Spec.Template.Spec.NodeSelector = m.OBProxy.Spec.NodeSelector
		deploy.Spec.Template.Spec.Affinity = m.OBProxy.Spec.Affinity
		deploy.Spec.Template.Spec.Tolerations = m.OBProxy.Spec.Tolerations

		return m.Client.Update(m.Ctx, deploy)
	})
}

func ScaleOBProxyDeployment(m *OBProxyManager) tasktypes.TaskError {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deploy, err := m.getOBProxyDeployment()
		if err != nil {
			return errors.Wrap(err, "get obproxy deployment")
		}
		if deploy == nil {
			return errors.New("deployment not found")
		}

		replicas := m.OBProxy.Spec.Replicas
		if deploy.Spec.Replicas != nil && *deploy.Spec.Replicas == replicas {
			return nil
		}

		deploy.Spec.Replicas = &replicas
		return m.Client.Update(m.Ctx, deploy)
	})
}

func DeleteOBProxyDeployment(m *OBProxyManager) tasktypes.TaskError {
	deploy, err := m.getOBProxyDeployment()
	if err != nil {
		return errors.Wrap(err, "get obproxy deployment")
	}
	if deploy == nil {
		m.Logger.Info("Deployment already deleted")
		return nil
	}

	err = m.Client.Delete(m.Ctx, deploy)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			m.Logger.Info("Deployment already deleted")
			return nil
		}
		return errors.Wrap(err, "delete deployment")
	}

	m.Logger.Info("Deleted Deployment", "deployment", deploy.Name)
	m.Recorder.Event(m.OBProxy, "DeleteDeployment", "", fmt.Sprintf("Deleted Deployment %s", deploy.Name))
	return nil
}

func DeleteOBProxyService(m *OBProxyManager) tasktypes.TaskError {
	svc, err := m.getOBProxyService()
	if err != nil {
		return errors.Wrap(err, "get obproxy service")
	}
	if svc == nil {
		m.Logger.Info("Service already deleted")
		return nil
	}

	err = m.Client.Delete(m.Ctx, svc)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			m.Logger.Info("Service already deleted")
			return nil
		}
		return errors.Wrap(err, "delete service")
	}

	m.Logger.Info("Deleted Service", "service", svc.Name)
	m.Recorder.Event(m.OBProxy, "DeleteService", "", fmt.Sprintf("Deleted Service %s", svc.Name))
	return nil
}

func DeleteOBProxyConfigMap(m *OBProxyManager) tasktypes.TaskError {
	cm, err := m.getOBProxyConfigMap()
	if err != nil {
		return errors.Wrap(err, "get obproxy configmap")
	}
	if cm == nil {
		m.Logger.Info("ConfigMap already deleted")
		return nil
	}

	err = m.Client.Delete(m.Ctx, cm)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			m.Logger.Info("ConfigMap already deleted")
			return nil
		}
		return errors.Wrap(err, "delete configmap")
	}

	m.Logger.Info("Deleted ConfigMap", "configmap", cm.Name)
	m.Recorder.Event(m.OBProxy, "DeleteConfigMap", "", fmt.Sprintf("Deleted ConfigMap %s", cm.Name))
	return nil
}

func DeleteOBProxySecrets(m *OBProxyManager) tasktypes.TaskError {
	proxyROSecretName := m.getProxyROSecretName()
	proxyROSecret := &corev1.Secret{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      proxyROSecretName,
	}, proxyROSecret)
	if err == nil {
		err = m.Client.Delete(m.Ctx, proxyROSecret)
		if err != nil && !kubeerrors.IsNotFound(err) {
			return errors.Wrap(err, "delete proxyRO secret")
		}
		m.Logger.Info("Deleted proxyRO secret", "secret", proxyROSecretName)
	}

	m.Recorder.Event(m.OBProxy, "DeleteSecrets", "", "Deleted OBProxy secrets")
	return nil
}

func (m *OBProxyManager) getRootServiceList() (string, string, error) {
	obcluster, err := m.getOBCluster()
	if err != nil {
		return "", "", errors.Wrap(err, "get obcluster")
	}

	observerList := &v1alpha1.OBServerList{}
	err = m.Client.List(m.Ctx, observerList,
		client.MatchingLabels{oceanbaseconst.LabelRefOBCluster: obcluster.Name},
		client.InNamespace(obcluster.Namespace),
	)
	if err != nil {
		return "", "", errors.Wrap(err, "list observers")
	}

	var rsList []string
	for _, observer := range observerList.Items {
		if observer.DeletionTimestamp != nil {
			continue
		}
		if observer.Status.Status == observerstatus.Running {
			rsList = append(rsList, fmt.Sprintf("%s:%d", observer.Status.GetConnectAddr(), oceanbaseconst.SqlPort))
		}
	}

	if len(rsList) > 0 {
		sort.Strings(rsList)
		return strings.Join(rsList, ";"), "k8s", nil
	}

	// Fallback: try to get from OceanBase parameter via SQL
	rs, err := m.getRootServiceListFromDB(obcluster)
	if err != nil {
		return "", "", err
	}
	return rs, "sql", nil
}

func (m *OBProxyManager) getRootServiceListFromDB(obcluster *v1alpha1.OBCluster) (string, error) {
	secretName := obcluster.Spec.UserSecrets.Root
	if secretName == "" {
		return "", errors.New("root secret not configured in obcluster")
	}

	secret := &corev1.Secret{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: obcluster.Namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		return "", errors.Wrap(err, "get root secret")
	}

	password := string(secret.Data["password"])

	observerList := &v1alpha1.OBServerList{}
	err = m.Client.List(m.Ctx, observerList,
		client.MatchingLabels{oceanbaseconst.LabelRefOBCluster: obcluster.Name},
		client.InNamespace(obcluster.Namespace),
	)
	if err != nil {
		return "", errors.Wrap(err, "list observers")
	}

	if len(observerList.Items) == 0 {
		return "", errors.New("no observers found")
	}

	for _, observer := range observerList.Items {
		if observer.Status.Status != observerstatus.Running {
			continue
		}

		address := observer.Status.GetConnectAddr()
		dataSource := connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, "root", "sys", password, oceanbaseconst.DefaultDatabase)
		manager, err := operation.GetOceanbaseOperationManager(dataSource)
		if err != nil {
			continue
		}

		parameters, err := manager.GetParameter(m.Ctx, "rootservice_list", nil)
		if err != nil {
			continue
		}

		if len(parameters) > 0 {
			rsList := parameters[0].Value
			rsList = strings.ReplaceAll(rsList, ":2882", "") // strip OB internal port suffix
			return rsList, nil
		}
	}

	return "", errors.New("failed to get rootservice_list from any observer")
}
