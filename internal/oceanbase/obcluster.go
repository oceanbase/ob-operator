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

package oceanbase

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/oceanbase/schema"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

const (
	PasswordKey = "password"
)

// TODO: generate random password
func generatePassword() string {
	return "pass"
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

func CreateSecretsForOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster, rootPass string) error {
	logger.Info("Create secrets for obcluster")
	err := createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.Root, rootPass)
	if err != nil {
		return errors.Wrap(err, "Create secret for user root")
	}
	err = createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.Monitor, generatePassword())
	if err != nil {
		return errors.Wrap(err, "Create secret for user monitor")
	}
	err = createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.ProxyRO, generatePassword())
	if err != nil {
		return errors.Wrap(err, "Create secret for user proxyro")
	}
	err = createPasswordSecret(ctx, obcluster.Namespace, obcluster.Spec.UserSecrets.Operator, generatePassword())
	if err != nil {
		return errors.Wrap(err, "Create secret for user operator")
	}
	return nil
}

func CreateOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster) error {
	logger.Infof("create obcluster with instance: %v", obcluster)
	client := client.GetClient()
	objectMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obcluster)
	if err != nil {
		return errors.Wrap(err, "Convert obcluster to unsturctured")
	}
	unstructuredObj := &unstructured.Unstructured{
		Object: objectMap,
	}
	unstructuredObj.SetGroupVersionKind(schema.OBClusterResKind)
	logger.Infof("create obcluster with unstructured: %v", unstructuredObj)
	_, err = client.DynamicClient.Resource(schema.OBClusterRes).Namespace(obcluster.Namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	return err
}

func UpdateOBCluster(ctx context.Context, obcluster *v1alpha1.OBCluster) error {
	client := client.GetClient()
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obcluster)
	if err != nil {
		return errors.Wrap(err, "Convert obcluster to unstructured")
	}
	_, err = client.DynamicClient.Resource(schema.OBClusterRes).Namespace(obcluster.Namespace).Update(ctx, &unstructured.Unstructured{
		Object: unstructuredObj,
	}, metav1.UpdateOptions{})
	return err
}

func GetOBCluster(ctx context.Context, namespace, name string) (*v1alpha1.OBCluster, error) {
	client := client.GetClient()
	obj, err := client.DynamicClient.Resource(schema.OBClusterRes).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var obcluster v1alpha1.OBCluster
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &obcluster)
	if err != nil {
		return nil, err
	}
	return &obcluster, nil
}

func DeleteOBCluster(ctx context.Context, namespace, name string) error {
	client := client.GetClient()
	err := client.DynamicClient.Resource(schema.OBClusterRes).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}

func ListAllOBClusters(ctx context.Context) (*v1alpha1.OBClusterList, error) {
	client := client.GetClient()
	obj, err := client.DynamicClient.Resource(schema.OBClusterRes).List(ctx, metav1.ListOptions{})
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
	obj, err := client.DynamicClient.Resource(schema.OBZoneRes).Namespace(obcluster.Namespace).List(ctx, metav1.ListOptions{
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
	logger.Infof("get observer list of obzone %s", obzone.Name)
	obj, err := client.DynamicClient.Resource(schema.OBServerRes).Namespace(obzone.Namespace).List(ctx, metav1.ListOptions{
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
