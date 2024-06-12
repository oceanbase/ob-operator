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

package obproxy

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/obproxy"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func getConfigMap(ctx context.Context, ns, name string) (*corev1.ConfigMap, error) {
	cm, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Get(ctx, cmPrefix+name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("ConfigMap not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy config map, err msg: " + err.Error())
	}
	return cm, nil
}

func createConfigMap(ctx context.Context, ns, name string, param *obproxy.CreateOBProxyParam) (*corev1.ConfigMap, error) {
	cmName := cmPrefix + name
	cmParam := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: ns,
			Labels: map[string]string{
				LabelOBProxy:            name,
				constant.LabelManagedBy: constant.DASHBOARD_APP_NAME,
			},
		},
		Data: map[string]string{},
	}
	for _, kv := range param.Parameters {
		cmParam.Data[strings.ToUpper(envPrefix+kv.Key)] = kv.Value
	}
	configMap, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Create(ctx, cmParam, metav1.CreateOptions{})
	if err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			return nil, httpErr.NewBadRequest("ConfigMap already exists")
		}
		return nil, httpErr.NewInternal("Failed to create obproxy config map, err msg: " + err.Error())
	}
	return configMap, nil
}

func updateConfigMap(ctx context.Context, ns, name string, param *obproxy.PatchOBProxyParam) (*corev1.ConfigMap, error) {
	cm, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Get(ctx, cmPrefix+name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("ConfigMap not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy config map, err msg: " + err.Error())
	}
	cm.Data = map[string]string{}
	for _, kv := range param.Parameters {
		cm.Data[strings.ToUpper(envPrefix+kv.Key)] = kv.Value
	}
	configMap, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to update obproxy config map, err msg: " + err.Error())
	}
	return configMap, nil
}

func doesParametersChanged(ctx context.Context, ns, name string, param *obproxy.PatchOBProxyParam) (bool, error) {
	cm, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Get(ctx, cmPrefix+name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return false, httpErr.NewNotFound("ConfigMap not found")
		}
		return false, httpErr.NewInternal("Failed to get obproxy config map, err msg: " + err.Error())
	}
	for _, kv := range param.Parameters {
		if val, ok := cm.Data[strings.ToUpper(envPrefix+kv.Key)]; !ok || val != kv.Value {
			return true, nil
		}
	}
	return false, nil
}

func deleteConfigMap(ctx context.Context, ns, name string) (*corev1.ConfigMap, error) {
	cm, err := client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Get(ctx, cmPrefix+name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("ConfigMap not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy config map, err msg: " + err.Error())
	}
	err = client.GetClient().ClientSet.CoreV1().ConfigMaps(ns).Delete(ctx, cmPrefix+name, metav1.DeleteOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to delete obproxy config map, err msg: " + err.Error())
	}
	return cm, nil
}
