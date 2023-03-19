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
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	ClientSet       *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
}

func NewClient() (*rest.Config, *Client) {
	client := new(Client)
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube", "config")); err != nil {
		config, _ := rest.InClusterConfig()
		config.QPS = 30
		config.Burst = 60
		client.ClientSet, _ = kubernetes.NewForConfig(config)
		client.DynamicClient, _ = dynamic.NewForConfig(config)
		client.DiscoveryClient, _ = discovery.NewDiscoveryClientForConfig(config)
		return config, client
	} else {
		filePath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, _ := clientcmd.BuildConfigFromFlags("", filePath)
		config.QPS = 30
		config.Burst = 60
		client.ClientSet, _ = kubernetes.NewForConfig(config)
		client.DynamicClient, _ = dynamic.NewForConfig(config)
		client.DiscoveryClient, _ = discovery.NewDiscoveryClientForConfig(config)
		return config, client
	}
}

func (client *Client) GetResource() ([]*metav1.APIGroup, []*metav1.APIResourceList) {
	group, source, _ := client.DiscoveryClient.ServerGroupsAndResources()
	return group, source
}

func (client *Client) GetKind(kind string) string {
	_, resourceList := client.GetResource()
	for _, list := range resourceList {
		for _, resource := range list.APIResources {
			if resource.Kind == kind {
				return resource.Name
			}
		}
	}
	return ""
}

func (client *Client) GetGVR(unStruct *unstructured.Unstructured) *schema.GroupVersionResource {
	gvk := unStruct.GroupVersionKind()
	kind := client.GetKind(gvk.Kind)
	return &schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: kind,
	}
}

func (client *Client) CreateObj(obj unstructured.Unstructured) error {
	gvr := client.GetGVR(&obj)
	_, err := client.DynamicClient.Resource(*gvr).Namespace(obj.GetNamespace()).Create(context.TODO(), &obj, metav1.CreateOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (client *Client) GetObj(obj unstructured.Unstructured) (interface{}, error) {
	gvr := client.GetGVR(&obj)
	res, err := client.DynamicClient.Resource(*gvr).Namespace(obj.GetNamespace()).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return res, err
	}
	return res, nil
}

func (client *Client) UpdateObj(obj unstructured.Unstructured) error {
	gvr := client.GetGVR(&obj)
	_, err := client.DynamicClient.Resource(*gvr).Namespace(obj.GetNamespace()).Update(context.TODO(), &obj, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (client *Client) DeleteObj(obj unstructured.Unstructured) {
	gvr := client.GetGVR(&obj)
	_ = client.DynamicClient.Resource(*gvr).Namespace(obj.GetNamespace()).Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
}
