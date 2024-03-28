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
	"k8s.io/apimachinery/pkg/util/intstr"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	podconst "github.com/oceanbase/ob-operator/internal/const/pod"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	observerstatus "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/status/server"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task-register $GOFILE

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
	obs, err := oceanbaseOperationManager.GetServer(serverInfo)
	if obs != nil {
		m.Logger.Info("OBServer already exists in obcluster")
		return nil
	}
	if err != nil {
		m.Logger.Error(err, "Get observer failed")
		return errors.Wrap(err, "Failed to get observer")
	}
	return oceanbaseOperationManager.AddServer(serverInfo)
}

func WaitOBClusterBootstrapped(m *OBServerManager) tasktypes.TaskError {
	for i := 0; i < oceanbaseconst.BootstrapTimeoutSeconds; i++ {
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

func CreateOBPod(m *OBServerManager) tasktypes.TaskError {
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
	// create pod
	observerPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            m.OBServer.Name,
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          m.OBServer.Labels,
			Annotations:     annotations,
		},
		Spec: observerPodSpec,
	}
	err = m.Client.Create(m.Ctx, observerPod)
	if err != nil {
		m.Logger.Error(err, "failed to create pod")
		return errors.Wrap(err, "failed to create pod")
	}
	m.Recorder.Event(m.OBServer, "CreatePod", "CreatePod", "Create observer pod")
	return nil
}

func CreateOBPVC(m *OBServerManager) tasktypes.TaskError {
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
				Labels:          m.OBServer.Labels,
			},
			Spec: m.generatePVCSpec(storageSpec),
		}
		err := m.Client.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create single pvc of observer")
		}
	} else {
		objectMeta := metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix),
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          m.OBServer.Labels,
		}
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       m.generatePVCSpec(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage),
		}
		err := m.Client.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create pvc of data file")
		}

		objectMeta = metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix),
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          m.OBServer.Labels,
		}
		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       m.generatePVCSpec(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage),
		}
		err = m.Client.Create(m.Ctx, pvc)
		if err != nil {
			return errors.Wrap(err, "Create pvc of data log")
		}

		objectMeta = metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix),
			Namespace:       m.OBServer.Namespace,
			OwnerReferences: ownerReferenceList,
			Labels:          m.OBServer.Labels,
		}
		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: objectMeta,
			Spec:       m.generatePVCSpec(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage),
		}
		err = m.Client.Create(m.Ctx, pvc)
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
	observer, err := operationManager.GetServer(observerInfo)
	if err != nil {
		return err
	}
	if observer != nil && observer.Status != "deleting" {
		if observer.Status == "deleting" {
			m.Logger.Info("OBServer is deleting", "observer", observerInfo.Ip)
		} else {
			m.Logger.Info("Need to delete observer")
			err = operationManager.DeleteServer(observerInfo)
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
	err = m.Client.Update(m.Ctx, observerPod)
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
	err = m.Client.Update(m.Ctx, observerPod)
	if err != nil {
		return errors.Wrapf(err, "Failed to update pod of observer %s", m.OBServer.Name)
	}
	return nil
}

func WaitOBServerPodReady(m *OBServerManager) tasktypes.TaskError {
	observerPodRestarted := false
	for i := 0; i < oceanbaseconst.DefaultStateWaitTimeout; i++ {
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
	if m.OBServer.SupportStaticIP() {
		return nil
	}
	m.Logger.Info("Wait for observer to be active in cluster")
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	active := false
	for i := 0; i < oceanbaseconst.DefaultStateWaitTimeout; i++ {
		operationManager, err := m.getOceanbaseOperationManager()
		if err != nil {
			return errors.Wrapf(err, "Get oceanbase operation manager failed")
		}
		observer, _ := operationManager.GetServer(observerInfo)
		if observer != nil {
			if observer.StartServiceTime > 0 && observer.Status == observerstatus.Active {
				active = true
				break
			}
		} else {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("OBServer is nil, check next time")
		}
		time.Sleep(time.Second)
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
	if m.OBServer.SupportStaticIP() {
		return nil
	}
	m.Logger.Info("Wait for observer to be deleted in cluster")
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.GetConnectAddr(),
		Port: oceanbaseconst.RpcPort,
	}
	deleted := false
	for i := 0; i < oceanbaseconst.ServerDeleteTimeoutSeconds; i++ {
		operationManager, err := m.getOceanbaseOperationManager()
		if err != nil {
			return errors.Wrapf(err, "Get oceanbase operation manager failed")
		}
		observer, err := operationManager.GetServer(observerInfo)
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
	err = m.Client.Delete(m.Ctx, pod)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete pod of observer %s", m.OBServer.Name)
	}

	return nil
}

func WaitForPodDeleted(m *OBServerManager) tasktypes.TaskError {
	m.Logger.Info("Wait for observer pod being deleted")
	for i := 0; i < oceanbaseconst.DefaultStateWaitTimeout; i++ {
		time.Sleep(time.Second)
		err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), &corev1.Pod{})
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
			err = m.Client.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix):
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size
			err = m.Client.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix):
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size
			err = m.Client.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		case m.OBServer.Name: // single pvc
			sum := resource.Quantity{}
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = sum
			err = m.Client.Update(m.Ctx, &pvc)
			if err != nil {
				return errors.Wrapf(err, "Failed to update pvc of observer %s", m.OBServer.Name)
			}
		}
	}
	return nil
}

func WaitForPVCResized(m *OBServerManager) tasktypes.TaskError {
outer:
	for i := 0; i < oceanbaseconst.DefaultStateWaitTimeout; i++ {
		time.Sleep(time.Second)

		observerPVC, err := m.getPVCs()
		if err != nil {
			return errors.Wrapf(err, "Failed to get pvc of observer %s", m.OBServer.Name)
		}
		for _, pvc := range observerPVC.Items {
			switch pvc.Name {
			case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.DataVolumeSuffix):
				if m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size.Cmp(pvc.Spec.Resources.Requests[corev1.ResourceStorage]) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.ClogVolumeSuffix):
				if m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Cmp(pvc.Spec.Resources.Requests[corev1.ResourceStorage]) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			case fmt.Sprintf("%s-%s", m.OBServer.Name, oceanbaseconst.LogVolumeSuffix):
				if m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size.Cmp(pvc.Spec.Resources.Requests[corev1.ResourceStorage]) != 0 {
					m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Data pvc not expanded", "pvc", pvc.Name)
					continue outer
				}
			case m.OBServer.Name:
				sum := resource.Quantity{}
				sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
				sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
				sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
				if sum.Cmp(pvc.Spec.Resources.Requests[corev1.ResourceStorage]) != 0 {
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
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.OBServer.Name,
				Namespace: m.OBServer.Namespace,
				Labels:    m.OBServer.Labels,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: m.OBServer.APIVersion,
					Kind:       m.OBServer.Kind,
					Name:       m.OBServer.Name,
					UID:        m.OBServer.GetUID(),
				}},
			},
			Spec: corev1.ServiceSpec{
				Selector: m.OBServer.Labels,
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
		err := m.Client.Create(m.Ctx, svc)
		if err != nil {
			return errors.Wrapf(err, "Failed to create observer service")
		}
	}
	return nil
}

func MountBackupVolume(_ *OBServerManager) tasktypes.TaskError {
	return nil
}

func WaitForBackupVolumeMounted(_ *OBServerManager) tasktypes.TaskError {
	return nil
}
