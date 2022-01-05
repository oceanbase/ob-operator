/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package resource

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
)

const (
	StatefulAppGroup    = "cloud.oceanbase.com"
	StatefulAppVersion  = "v1"
	StatefulAppKind     = "StatefulApp"
	StatefulAppResource = "statefulapps"
)

var (
	StatefulappRes = schema.GroupVersionResource{
		Group:    StatefulAppGroup,
		Version:  StatefulAppVersion,
		Resource: StatefulAppResource,
	}
)

func (client *Client) GetStatefulappInstance(namespace, name string) (cloudv1.StatefulApp, error) {
	var instance cloudv1.StatefulApp
	obj, err := client.DynamicClient.Resource(StatefulappRes).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return instance, err
	}
	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &instance)
	return instance, nil
}

func (client *Client) UpdateStatefulappInstance(obj unstructured.Unstructured) error {
	oldObj, _ := client.GetObj(obj)
	obj.SetResourceVersion(oldObj.(*unstructured.Unstructured).GetResourceVersion())
	err := client.UpdateObj(obj)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (client *Client) JudgeStatefulappInstanceIsReadyByObj(namespace, name string) bool {
	instance, err := client.GetStatefulappInstance(namespace, name)
	if err == nil {
		return testconverter.IsStatefulappInstanceReady(instance)
	}
	return false
}
