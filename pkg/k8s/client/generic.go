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

package client

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sResourceClient[T runtimeclient.Object] interface {
	List(ctx context.Context, namespace string, list runtimeclient.ObjectList, opts v1.ListOptions) error
	Get(ctx context.Context, namespace, name string, opts v1.GetOptions) (T, error)
	Create(ctx context.Context, obj T, opts v1.CreateOptions) (T, error)
	Update(ctx context.Context, obj T, opts v1.UpdateOptions) (T, error)
	Delete(ctx context.Context, namespace, name string, opts v1.DeleteOptions) error
}

func NewDynamicResourceClient[T runtimeclient.Object](gvr schema.GroupVersionResource, kind string) K8sResourceClient[T] {
	client := GetClient()
	return dynamicResourceClient[T]{
		client: client.DynamicClient.Resource(gvr),
		gvr:    gvr,
		kind:   kind,
	}
}

type dynamicResourceClient[T runtimeclient.Object] struct {
	client dynamic.NamespaceableResourceInterface
	gvr    schema.GroupVersionResource
	kind   string
}

// List lists objects with the given options, store the result in the given list object
func (c dynamicResourceClient[T]) List(ctx context.Context, namespace string, list runtimeclient.ObjectList, opts v1.ListOptions) error {
	if list == nil {
		return errors.New("target list object is nil")
	}
	obj, err := c.client.Namespace(namespace).List(ctx, opts)
	if err != nil {
		return err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), list)
	if err != nil {
		return err
	}
	return nil
}

// Get gets the object with the given namespace, name and options
func (c dynamicResourceClient[T]) Get(ctx context.Context, namespace, name string, opts v1.GetOptions) (T, error) {
	obj, err := c.client.Namespace(namespace).Get(ctx, name, opts)
	if err != nil {
		return *new(T), err
	}
	var item T
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &item)
	if err != nil {
		return *new(T), err
	}
	return item, nil
}

func (c dynamicResourceClient[T]) Create(ctx context.Context, obj T, opts v1.CreateOptions) (T, error) {
	if obj.GetName() == "" {
		return *new(T), fmt.Errorf("the name of %s is empty", c.kind)
	}
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return *new(T), err
	}

	unstructuredObj := &unstructured.Unstructured{
		Object: objMap,
	}
	unstructuredObj.SetGroupVersionKind(c.gvr.GroupVersion().WithKind(c.kind))
	res, err := c.client.Namespace(obj.GetNamespace()).Create(ctx, unstructuredObj, opts)
	if err != nil {
		return *new(T), err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.UnstructuredContent(), obj)
	if err != nil {
		return *new(T), err
	}
	return obj, nil
}

func (c dynamicResourceClient[T]) Update(ctx context.Context, obj T, opts v1.UpdateOptions) (T, error) {
	if obj.GetName() == "" {
		return *new(T), fmt.Errorf("the name of %s is empty", c.kind)
	}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return *new(T), err
	}

	res, err := c.client.Namespace(obj.GetNamespace()).Update(ctx, &unstructured.Unstructured{
		Object: unstructuredObj,
	}, opts)
	if err != nil {
		return *new(T), err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.UnstructuredContent(), obj)
	if err != nil {
		return *new(T), err
	}
	return obj, nil
}

func (c dynamicResourceClient[T]) Delete(ctx context.Context, namespace, name string, opts v1.DeleteOptions) error {
	return c.client.Namespace(namespace).Delete(ctx, name, opts)
}
