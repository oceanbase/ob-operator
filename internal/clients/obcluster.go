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

package clients

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients/schema"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

const (
	PasswordKey = "password"
)

func generatePassword() string {
	return rand.String(16)
}

func createPasswordSecret(ctx context.Context, namespace, name, password string) error {
	client := client.GetClient()
	stringData := make(map[string]string)
	stringData[PasswordKey] = password
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Type:       "Opaque",
		StringData: stringData,
	}
	_, err := client.ClientSet.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func CreateSecretsForOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster, param *param.CreateOBClusterParam) error {
	logger.Info("Create secrets for obcluster")
	err := createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.Root, param.RootPassword)
	if err != nil {
		return errors.Wrap(err, "Create secret for user root")
	}
	proxyroPassword := param.ProxyroPassword
	if proxyroPassword == "" {
		proxyroPassword = generatePassword()
	}
	err = createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.ProxyRO, proxyroPassword)
	if err != nil {
		return errors.Wrap(err, "Create secret for user proxyro")
	}
	err = createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.Monitor, generatePassword())
	if err != nil {
		return errors.Wrap(err, "Create secret for user monitor")
	}
	err = createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.Operator, generatePassword())
	if err != nil {
		return errors.Wrap(err, "Create secret for user operator")
	}
	return nil
}

func CreateOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster) (*v1alpha1.OBCluster, error) {
	return ClusterClient.Create(ctx, obcluster, metav1.CreateOptions{})
}

func UpdateOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster) (*v1alpha1.OBCluster, error) {
	return ClusterClient.Update(ctx, obcluster, metav1.UpdateOptions{})
}

func GetOBCluster(ctx context.Context, namespace, name string) (*v1alpha1.OBCluster, error) {
	return ClusterClient.Get(ctx, namespace, name, metav1.GetOptions{})
}

func DeleteOBCluster(ctx context.Context, namespace, name string) error {
	return ClusterClient.Delete(ctx, namespace, name, metav1.DeleteOptions{})
}

func CreateOBClusterOperation(ctx context.Context, obclusterOperation *v1alpha1.OBClusterOperation) (*v1alpha1.OBClusterOperation, error) {
	return ClusterOperationClient.Create(ctx, obclusterOperation, metav1.CreateOptions{})
}

func GetOBClusterOperations(ctx context.Context, obcluster *v1alpha1.OBCluster) (*v1alpha1.OBClusterOperationList, error) {
	client := client.GetClient()
	var obclustetOperationList v1alpha1.OBClusterOperationList
	obj, err := client.DynamicClient.Resource(schema.OBClusterOperationGVR).Namespace(obcluster.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", oceanbaseconst.LabelRefOBClusterOp, obcluster.Name),
	})
	if err != nil {
		return nil, errors.Wrap(err, "List obcluster operations")
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &obclustetOperationList)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured to obcluster list")
	}
	return &obclustetOperationList, nil
}

func ListAllOBClusters(ctx context.Context) (*v1alpha1.OBClusterList, error) {
	client := client.GetClient()
	obj, err := client.DynamicClient.Resource(schema.OBClusterGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var obclusterList v1alpha1.OBClusterList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &obclusterList)
	if err != nil {
		return nil, err
	}
	return &obclusterList, nil
}

func ListAllOBServers(ctx context.Context) (*v1alpha1.OBClusterList, error) {
	client := client.GetClient()
	obj, err := client.DynamicClient.Resource(schema.OBClusterGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var obclusterList v1alpha1.OBClusterList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &obclusterList)
	if err != nil {
		return nil, err
	}
	return &obclusterList, nil
}

func ListOBZonesOfOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster) (*v1alpha1.OBZoneList, error) {
	client := client.GetClient()
	var obzoneList v1alpha1.OBZoneList
	obj, err := client.DynamicClient.Resource(schema.OBZoneGVR).Namespace(obcluster.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", oceanbaseconst.LabelRefOBCluster, obcluster.Name),
	})
	if err != nil {
		return nil, errors.Wrap(err, "List obzones")
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &obzoneList)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured to obzone list")
	}
	return &obzoneList, nil
}

func ListOBServersOfOBZone(ctx context.Context, obzone *v1alpha1.OBZone) (*v1alpha1.OBServerList, error) {
	client := client.GetClient()
	var observerList v1alpha1.OBServerList
	obj, err := client.DynamicClient.Resource(schema.OBServerGVR).Namespace(obzone.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", oceanbaseconst.LabelRefOBZone, obzone.Name),
	})
	if err != nil {
		return nil, errors.Wrap(err, "List observers")
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &observerList)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured to observer list")
	}
	return &observerList, nil
}

func GetPodOfOBServer(ctx context.Context, observer *v1alpha1.OBServer) (*corev1.Pod, error) {
	client := client.GetClient()
	// pod name is the same as observer name
	pod, err := client.ClientSet.CoreV1().Pods(observer.Namespace).Get(ctx, observer.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get pod")
	}
	return pod, nil
}

func ListOBParametersOfOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster) (*v1alpha1.OBParameterList, error) {
	client := client.GetClient()
	var obparameterList v1alpha1.OBParameterList
	obj, err := client.DynamicClient.Resource(schema.OBParameterGVR).Namespace(obcluster.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", oceanbaseconst.LabelRefUID, string(obcluster.GetUID())),
	})
	if err != nil {
		return nil, errors.Wrap(err, "List obparameters")
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &obparameterList)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured to obparameter list")
	}
	return &obparameterList, nil
}
