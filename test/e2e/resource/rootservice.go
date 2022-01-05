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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

const (
	RootServiceGroup    = "cloud.oceanbase.com"
	RootServiceVersion  = "v1"
	RootServiceKind     = "RootService"
	RootServiceResource = "rootservices"
)

var (
	RootServiceRes = schema.GroupVersionResource{
		Group:    RootServiceGroup,
		Version:  RootServiceVersion,
		Resource: RootServiceResource,
	}
)

func (client *Client) GetRootService(namespace, name string) (cloudv1.RootService, error) {
	var instance cloudv1.RootService
	obj, err := client.DynamicClient.Resource(RootServiceRes).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return instance, err
	}
	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &instance)
	return instance, nil
}
