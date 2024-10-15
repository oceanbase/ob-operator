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
	"sync"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
)

type Client struct {
	ClientSet     *kubernetes.Clientset
	DynamicClient dynamic.Interface
	MetaClient    metadata.Interface
	config        *rest.Config
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

func MustGetClient(config *rest.Config) *Client {
	client, err := getClientFromConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func (c *Client) GetConfig() *rest.Config {
	return c.config
}

func getClientFromConfig(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset")
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dynamic client")
	}
	metaClient, err := metadata.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create meta client")
	}
	return &Client{
		ClientSet:     clientset,
		DynamicClient: dynamicClient,
		MetaClient:    metaClient,
		config:        config,
	}, nil
}

func GetClientFromBytes(kubeconfig []byte) (*Client, error) {
	config, err := GetConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}
	return getClientFromConfig(config)
}
