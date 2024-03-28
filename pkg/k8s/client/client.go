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
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	ClientSet       *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
	config          *rest.Config
}

var client *Client

var clientOnce sync.Once

// maybe need to create client each time
func GetClient() *Client {
	clientOnce.Do(
		func() {
			var conf *rest.Config
			if _, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST"); exist {
				conf = MustGetConfigInCluster()
			} else {
				conf = MustGetConfigOutsideCluster()
			}
			client = MustGetClient(conf)
		},
	)
	return client
}

func MustGetConfigInCluster() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return config
}

func MustGetConfigOutsideCluster() *rest.Config {
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	fmt.Println("home:", home)
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	var configPath string
	configPathEnv, exist := os.LookupEnv("KUBECONFIG")
	if exist && configPathEnv != "" {
		configPath = configPathEnv
	} else {
		home := homedir.HomeDir()
		configPath = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func MustGetClient(config *rest.Config) *Client {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Client{
		ClientSet:       clientset,
		DynamicClient:   dynamicClient,
		DiscoveryClient: discoveryClient,
		config:          config,
	}
}

func (c *Client) GetConfig() *rest.Config {
	return c.config
}
