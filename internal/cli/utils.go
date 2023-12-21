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

package cli

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func objToUns(obj client.Object) (*unstructured.Unstructured, *meta.RESTMapping, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}
	unsContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, err
	}
	uns := &unstructured.Unstructured{Object: unsContent}
	return uns, mapping, nil
}

func createObj(ctx context.Context, obj client.Object) (*unstructured.Unstructured, error) {
	uns, mapping, err := objToUns(obj)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace()).Create(ctx, uns, metav1.CreateOptions{})
}

func updateObj(ctx context.Context, obj client.Object) (*unstructured.Unstructured, error) {
	uns, mapping, err := objToUns(obj)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace()).Update(ctx, uns, metav1.UpdateOptions{})
}

func deleteObj(ctx context.Context, obj client.Object) error {
	_, mapping, err := objToUns(obj)
	if err != nil {
		return err
	}
	return dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace()).Delete(ctx, obj.GetName(), metav1.DeleteOptions{})
}

func getObj(ctx context.Context, obj client.Object) (*unstructured.Unstructured, error) {
	_, mapping, err := objToUns(obj)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})
}
