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

package observer

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	podconst "github.com/oceanbase/ob-operator/internal/const/pod"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	observerstatus "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/status/server"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*OBServerManager]()

func WaitOBServerReady(m *OBServerManager) tasktypes.TaskError {
	for i := 0; i < podconst.ReadyTimeoutSeconds; i++ {
		observer, err := m.getOBServer()
		if err != nil {
			return errors.Wrap(err, "Get observer from K8s")
		}
		if observer.Status.Ready {
			m.Logger.Info("Pod is ready")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("Timeout to wait pod ready")
}

func AddServer(m *OBServerManager) tasktypes.TaskError {
	mode, modeAnnoExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist && mode == oceanbaseconst.ModeStandalone {
		m.Recorder.Event(m.OBServer, "SkipAddServer", "AddServer", "Skip add server in standalone mode")
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Skip add server in standalone mode")
		return nil
	}
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get oceanbase operation manager failed")
		return errors.Wrap(err, "Get oceanbase operation manager")
	}
	serverInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	obs, err := oceanbaseOperationManager.GetServer(m.Ctx, serverInfo)
	if obs != nil {
		m.Logger.Info("OBServer already exists in obcluster")
		return nil
	}
	if err != nil {
		m.Logger.Error(err, "Get observer failed")
		return errors.Wrap(err, "Failed to get observer")
	}
	return oceanbaseOperationManager.AddServer(m.Ctx, serverInfo)
}

func WaitOBClusterBootstrapped(m *OBServerManager) tasktypes.TaskError {
	for i := 0; i < obcfg.GetConfig().Time.BootstrapTimeoutSeconds; i++ {
		obcluster, err := m.getOBCluster()
		if err != nil {
			return errors.Wrap(err, "Get obcluster from K8s")
		}
		if obcluster.Status.Status == clusterstatus.Bootstrapped {
			m.Logger.Info("OBCluster bootstrapped")
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("Timeout to wait obcluster bootstrapped")
}

func CreateOBServerPod(m *OBServerManager) tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Create observer pod")
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster from K8s")
	}
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	ownerReference := metav1.OwnerReference{
		APIVersion: m.OBServer.APIVersion,
		Kind:       m.OBServer.Kind,
		Name:       m.OBServer.Name,
		UID:        m.OBServer.GetUID(),
	}
	annotations := m.generateStaticIpAnnotation()
	ownerReferenceList = append(ownerReferenceList, ownerReference)
	observerPodSpec := m.createOBPodSpec(obcluster)
	podLabels := m.OBServer.Labels
	podLabels[oceanbaseconst.LabelRefUID] = string(m.OBServer.UID)
	podLabels[oceanbaseconst.LabelOBServerUID] = string(m.OBServer.UID) // For compatibility with old version
	podLabels[oceanbaseconst.LabelRefOBServer] = string(m.OBServer.Name)

	podFields := m.OBServer.Spec.OBServerTemplate.PodFields
	if podFields != nil {
		varsReplacer := m.getVarsReplacer(obcluster)
		if podFields.HostName != nil && *podFields.HostName != "" {
			observerPodSpec.Hostname = varsReplacer.Replace(*podFields.HostName)
		}
		if podFields.Subdomain != nil && *podFields.Subdomain != "" {
			observerPodSpec.Subdomain = varsReplacer.Replace(*podFields.Subdomain)
		}
		for k := range podFields.Labels {
			if _, exist := podLabels[k]; !exist {
				podLabels[k] = varsReplacer.Replace(podFields.Labels[k])
			}
		}
		for k := range podFields.Annotations {
			if _, exist := annotations[k]; !exist {
				annotations[k] = varsReplacer.Replace(podFields.Annotations[k])
			}
		}
	}

	// create pod
	observerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            m.OBServer.Name,
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          podLabels,
			Annotations:     annotations,
		},
		Spec: observerPodSpec,
	}
	err = m.K8sResClient.Create(m.Ctx, observerPod)
	if err != nil {
		m.Logger.Error(err, "failed to create pod")
		return errors.Wrap(err, "failed to create pod")
	}
	m.Recorder.Event(m.OBServer, "CreatePod", "CreatePod", "Create observer pod")
	return nil
}

func CreateOBServerPVC(m *OBServerManager) tasktypes.TaskError {
	ownerReferenceList := make([]metav1.OwnerReference, 0)
	sepVolumeAnnoVal, sepVolumeAnnoExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsIndependentPVCLifecycle)
	if !sepVolumeAnnoExist || sepVolumeAnnoVal != "true" {
		ownerReference := metav1.OwnerReference{
			APIVersion: m.OBServer.APIVersion,
			Kind:       m.OBServer.Kind,
			Name:       m.OBServer.Name,
			UID:        m.OBServer.GetUID(),
		}
		ownerReferenceList = append(ownerReferenceList, ownerReference)
	}
	singlePvcAnnoVal, singlePvcExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsSinglePVC)
	pvcLabels := m.OBServer.Labels
	pvcLabels[oceanbaseconst.LabelRefUID] = string(m.OBServer.UID)
	pvcLabels[oceanbaseconst.LabelRefOBServer] = string(m.OBServer.Name)

	if singlePvcExist && singlePvcAnnoVal == "true" {
		sumQuantity := resource.Quantity{}
		sumQuantity.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
		sumQuantity.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
		sumQuantity.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
		storageSpec := &apitypes.StorageSpec{
			StorageClass: m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.StorageClass,
			Size:         sumQuantity,
		}
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:            m.OBServer.Name,
				Namespace:       m.OBServer.Namespace,
				OwnerReferences: ownerReferenceList,
				Labels:          pvcLabels,
			},
			Spec: m.generatePVCSpec(storageSpec),
		}
		err := m.K8sResClient.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create single pvc of observer")
		}
	} else {
		objectMeta := metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix),
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          pvcLabels,
		}
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       m.generatePVCSpec(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage),
		}
		err := m.K8sResClient.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create pvc of data file")
		}

		objectMeta = metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix),
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          pvcLabels,
		}
		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       m.generatePVCSpec(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage),
		}
		err = m.K8sResClient.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create pvc of data log")
		}

		objectMeta = metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix),
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          pvcLabels,
		}
		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       m.generatePVCSpec(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage),
		}
		err = m.K8sResClient.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create pvc of log")
		}
	}

	return nil
}

func DeleteOBServerInCluster(m *OBServerManager) tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Delete observer in cluster")
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrapf(err, "Get oceanbase operation manager failed")
	}
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	observer, err := operationManager.GetServer(m.Ctx, observerInfo)
	if err != nil {
		return err
	}
	if observer != nil && observer.Status != "deleting" {
		if observer.Status == "deleting" {
			m.Logger.Info("OBServer is deleting", "observer", observerInfo.Ip)
		} else {
			m.Logger.Info("Need to delete observer")
			err = operationManager.DeleteServer(m.Ctx, observerInfo)
			if err != nil {
				return errors.Wrapf(err, "Failed to delete observer %s", observerInfo.Ip)
			}
		}
	} else {
		m.Logger.Info("OBServer already deleted", "observer", observerInfo.Ip)
	}
	return nil
}

func AnnotateOBServerPod(m *OBServerManager) tasktypes.TaskError {
	observerPod, err := m.getPod()
	if err != nil {
		return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
	}
	if m.OBServer.Status.CNI == oceanbaseconst.CNICalico {
		m.Logger.Info("Update pod annotation, cni is calico")
		observerPod.Annotations[oceanbaseconst.AnnotationCalicoIpAddrs] = fmt.Sprintf("[\"%s\"]", m.OBServer.Status.PodIp)
	}
	if m.OBServer.Status.CNI == oceanbaseconst.CNIKubeOvn {
		m.Logger.Info("Update pod annotation, cni is kube-ovn")
		observerPod.Annotations[oceanbaseconst.AnnotationKubeOvnIpAddrs] = m.OBServer.Status.PodIp
	}
	err = m.K8sResClient.Update(m.Ctx, observerPod)
	if err != nil {
		return errors.Wrapf(err, "Failed to update pod annotation of observer %s", m.OBServer.Name)
	}
	return nil
}

func UpgradeOBServerImage(m *OBServerManager) tasktypes.TaskError {
	observerPod, err := m.getPod()
	if err != nil {
		return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
	}
	for idx, container := range observerPod.Spec.Containers {
		if container.Name == oceanbaseconst.ContainerName {
			observerPod.Spec.Containers[idx].Image = m.OBServer.Spec.OBServerTemplate.Image
			break
		}
	}
	err = m.K8sResClient.Update(m.Ctx, observerPod)
	if err != nil {
		return errors.Wrapf(err, "Failed to update pod of observer %s", m.OBServer.Name)
	}
	return nil
}

func WaitOBServerPodReady(m *OBServerManager) tasktypes.TaskError {
	observerPodRestarted := false
	for i := 0; i < obcfg.GetConfig().Time.DefaultStateWaitTimeout; i++ {
		observerPod, err := m.getPod()
		if err != nil {
			return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
		}
		for _, containerStatus := range observerPod.Status.ContainerStatuses {
			if containerStatus.Name != oceanbaseconst.ContainerName {
				continue
			}
			if containerStatus.Ready && containerStatus.Image == m.OBServer.Spec.OBServerTemplate.Image {
				observerPodRestarted = true
			}
		}
		if observerPodRestarted {
			m.Logger.Info("OBServer pod restarted")
			break
		}
		time.Sleep(time.Second)
	}
	if !observerPodRestarted {
		return errors.Errorf("observer %s pod still not restart when timeout", m.OBServer.Name)
	}
	return nil
}

func WaitOBServerActiveInCluster(m *OBServerManager) tasktypes.TaskError {
	m.Logger.Info("Wait for observer to be active in cluster")
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	if m.OBServer.Annotations[oceanbaseconst.AnnotationsMode] == oceanbaseconst.ModeStandalone {
		observerInfo.Ip = "127.0.0.1"
	}
	active := false
	for i := 0; i < obcfg.GetConfig().Time.DefaultStateWaitTimeout; i++ {
		time.Sleep(time.Second)
		operationManager, err := m.getOceanbaseOperationManager()
		if err != nil {
			m.Logger.Error(err, "Get oceanbase operation manager failed")
			continue
		}
		observer, _ := operationManager.GetServer(m.Ctx, observerInfo)
		if observer != nil {
			if observer.StartServiceTime > 0 && observer.Status == observerstatus.Active {
				active = true
				break
			}
		} else {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("OBServer is nil, check next time")
		}
	}
	if !active {
		m.Logger.Info("Wait for observer to become active, timeout")
		return errors.Errorf("Wait observer %s active timeout", observerInfo.Ip)
	}
	m.Logger.Info("OBServer becomes active", "observer", observerInfo)
	m.Recorder.Event(m.OBServer, "OBServerBecomesActive", "OBServerBecomesActive", "OBServer becomes active")
	return nil
}

func WaitOBServerDeletedInCluster(m *OBServerManager) tasktypes.TaskError {
	m.Logger.Info("Wait for observer to be deleted in cluster")
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	deleted := false
	for i := 0; i < obcfg.GetConfig().Time.ServerDeleteTimeoutSeconds; i++ {
		operationManager, err := m.getOceanbaseOperationManager()
		if err != nil {
			return errors.Wrapf(err, "Get oceanbase operation manager failed")
		}
		observer, err := operationManager.GetServer(m.Ctx, observerInfo)
		if observer == nil && err == nil {
			m.Logger.Info("OBServer deleted")
			deleted = true
			break
		} else if err != nil {
			m.Logger.Error(err, "Query observer info failed")
		}
		time.Sleep(time.Second)
	}
	if !deleted {
		m.Logger.Info("Wait observer deleted timeout")
		return errors.Errorf("Wait observer %s deleted timeout", observerInfo.Ip)
	}
	m.Logger.Info("OBServer was deleted", "observer", observerInfo)
	m.Recorder.Event(m.OBServer, "OBServerDeleted", "OBServerDeleted", "OBServer was deleted")
	return nil
}

func DeletePod(m *OBServerManager) tasktypes.TaskError {
	m.Logger.Info("Delete observer pod")
	pod, err := m.getPod()
	if err != nil {
		return errors.Wrapf(err, "Failed to get pod of observer %s", m.OBServer.Name)
	}
	err = m.K8sResClient.Delete(m.Ctx, pod)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete pod of observer %s", m.OBServer.Name)
	}

	return nil
}

func WaitForPodDeleted(m *OBServerManager) tasktypes.TaskError {
	m.Logger.Info("Wait for observer pod being deleted")
	for i := 0; i < obcfg.GetConfig().Time.DefaultStateWaitTimeout; i++ {
		time.Sleep(time.Second)
		err := m.K8sResClient.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), &corev1.Pod{})
		if err != nil && kubeerrors.IsNotFound(err) {
			return nil
		}
	}
	return errors.New("Timeout to wait for pod being deleted")
}

func ExpandPVC(m *OBServerManager) tasktypes.TaskError {
	observerPVC, err := m.getPVCs()
	if err != nil {
		return errors.Wrapf(err, "Failed to get pvc list of observer %s", m.OBServer.Name)
	}

	for _, pvc := range observerPVC.Items {
		switch pvc.Name {
		case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix):
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size
			err = m.K8sResClient.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix):
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size
			err = m.K8sResClient.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix):
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size
			err = m.K8sResClient.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		case m.OBServer.Name: // single pvc
			sum := resource.Quantity{}
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = sum
			err = m.K8sResClient.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		}
	}
	return nil
}

func WaitForPvcResized(m *OBServerManager) tasktypes.TaskError {
outer:
	for i := 0; i < obcfg.GetConfig().Time.DefaultStateWaitTimeout; i++ {
		time.Sleep(time.Second)

		observerPVC, err := m.getPVCs()
		if err != nil {
			return errors.Wrapf(err, "Failed to get pvc of observer %s", m.OBServer.Name)
		}
		serverStorage := m.OBServer.Spec.OBServerTemplate.Storage
		for _, pvc := range observerPVC.Items {
			pvcSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			switch pvc.Name {
			case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix):
				if serverStorage.DataStorage.Size.Cmp(pvcSize) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix):
				if serverStorage.RedoLogStorage.Size.Cmp(pvcSize) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix):
				if serverStorage.LogStorage.Size.Cmp(pvcSize) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			case m.OBServer.Name:
				sum := resource.Quantity{}
				sum.Add(serverStorage.DataStorage.Size)
				sum.Add(serverStorage.RedoLogStorage.Size)
				sum.Add(serverStorage.LogStorage.Size)
				if sum.Cmp(pvcSize) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			}
		}
		// all pvc expanded
		return nil
	}
	return errors.Errorf("Timeout to wait for pvc expanded")
}

func CreateOBServerSvc(m *OBServerManager) tasktypes.TaskError {
	mode, modeAnnoExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if modeAnnoExist && mode == oceanbaseconst.ModeService {
		m.Logger.Info("Create observer service")
		svcLabels := m.OBServer.Labels
		svcLabels[oceanbaseconst.LabelRefUID] = string(m.OBServer.UID)
		svcLabels[oceanbaseconst.LabelRefOBServer] = string(m.OBServer.Name)
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.OBServer.Name,
				Namespace: m.OBServer.Namespace,
				Labels:    svcLabels,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: m.OBServer.APIVersion,
					Kind:       m.OBServer.Kind,
					Name:       m.OBServer.Name,
					UID:        m.OBServer.GetUID(),
				}},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					oceanbaseconst.LabelOBServerUID: string(m.OBServer.UID),
				},
				Ports: []corev1.ServicePort{{
					Name:       "sql",
					Port:       oceanbaseconst.SqlPort,
					TargetPort: intstr.IntOrString{IntVal: oceanbaseconst.SqlPort},
				}, {
					Name:       "rpc",
					Port:       oceanbaseconst.RpcPort,
					TargetPort: intstr.IntOrString{IntVal: oceanbaseconst.RpcPort},
				}},
			},
		}
		err := m.K8sResClient.Create(m.Ctx, svc)
		if err != nil {
			return errors.Wrapf(err, "Failed to create observer service")
		}
	}
	return nil
}

func CheckAndCreateNs(m *OBServerManager) tasktypes.TaskError {
	if !m.OBServer.InMasterK8s() {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: m.OBServer.Namespace,
			},
		}
		err := m.K8sResClient.Get(m.Ctx, types.NamespacedName{Name: m.OBServer.Namespace}, ns)
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				m.Logger.Info("Create namespace", "namespace", m.OBServer.Namespace, "k8sCluster", m.OBServer.Spec.K8sCluster)
				err = m.K8sResClient.Create(m.Ctx, ns)
				if err != nil {
					return errors.Wrapf(err, "Failed to create namespace %s with credential of k8s cluster %s", m.OBServer.Namespace, m.OBServer.Spec.K8sCluster)
				}
			} else {
				return errors.Wrapf(err, "Failed to get namespace %s with credential of k8s cluster %s", m.OBServer.Namespace, m.OBServer.Spec.K8sCluster)
			}
		}
	}
	return nil
}

func CleanOwnedResources(m *OBServerManager) tasktypes.TaskError {
	if !m.OBServer.InMasterK8s() {
		err := m.cleanWorkerK8sResource()
		if err != nil {
			m.Logger.Error(err, "Failed to clean worker k8s resources",
				"observer", m.OBServer.Name,
				"namespace", m.OBServer.Namespace,
			)
		}
	}
	return nil
}
