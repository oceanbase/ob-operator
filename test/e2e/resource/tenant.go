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
	TenantGroup    = "cloud.oceanbase.com"
	TenantVersion  = "v1"
	TenantKind     = "Tenant"
	TenantResource = "tenants"
)

var (
	TenantRes = schema.GroupVersionResource{
		Group:    TenantGroup,
		Version:  TenantVersion,
		Resource: TenantResource,
	}
)

func (client *Client) GetTenantInstance(namespace, name string) (cloudv1.Tenant, error) {
	var instance cloudv1.Tenant
	obj, err := client.DynamicClient.Resource(TenantRes).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return instance, err
	}
	_ = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &instance)
	return instance, nil
}

func (client *Client) GetTenantStatus(namespace, name string) string {
	instance, err := client.GetTenantInstance(namespace, name)
	if err == nil {
		return instance.Status.Status
	}
	return ""
}

func (client *Client) UpdateTenantInstance(obj unstructured.Unstructured) error {
	oldObj, _ := client.GetObj(obj)
	obj.SetResourceVersion(oldObj.(*unstructured.Unstructured).GetResourceVersion())
	err := client.UpdateObj(obj)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (client *Client) JudgeTenantInstanceIsRunningByObj(namespace, name string) bool {
	instance, err := client.GetTenantInstance(namespace, name)
	if err == nil {
		return testconverter.IsTenantInstanceRunning(instance)
	}
	return false
}

func (client *Client) JudgeTenantResourceUnitIsMatched(namespace, name string) bool {
	instance, err := client.GetTenantInstance(namespace, name)
	if err == nil {
		return testconverter.IsTenantResourceUnitMatched(instance)
	}
	return false
}

func (client *Client) JudgeTenantPrimaryZoneIsMatched(namespace, name string) bool {
	instance, err := client.GetTenantInstance(namespace, name)
	if err == nil {
		return testconverter.IsTenantPrimaryZoneMatched(instance)
	}
	return false
}

func (client *Client) JudgeTenantUnitNumIsMatched(namespace, name string) bool {
	instance, err := client.GetTenantInstance(namespace, name)
	if err == nil {
		return testconverter.IsTenantUnitNumMatched(instance)
	}
	return false
}

func (client *Client) JudgeTenantLocalityIsMatched(namespace, name string) bool {
	instance, err := client.GetTenantInstance(namespace, name)
	if err == nil {
		return testconverter.IsTenantLocalityMatched(instance)
	}
	return false
}
