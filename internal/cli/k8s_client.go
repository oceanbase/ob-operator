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
	"log"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var kubeconfig *string
var restConfig *rest.Config
var clientset *kubernetes.Clientset
var dynamicClient dynamic.Interface
var mapper meta.RESTMapper

func initClients() {
	home := homedir.HomeDir()
	kubeconfig = rootCmd.PersistentFlags().StringP("kubecfonfig", "c", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")

	var err error
	restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Println(err.Error())
		return
	}
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Println(err.Error())
		return
	}
	dynamicClient = dynamic.NewForConfigOrDie(restConfig)
	groupResources, err := restmapper.GetAPIGroupResources(clientset.DiscoveryClient)
	if err != nil {
		log.Println(err.Error())
		return
	}
	mapper = restmapper.NewDiscoveryRESTMapper(groupResources)
}
