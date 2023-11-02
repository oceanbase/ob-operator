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

package telemetry

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/oceanbase/ob-operator/pkg/telemetry/models"
)

type hostMetrics struct {
	IPHashes []string         `json:"ipHashes"`
	K8sNodes []models.K8sNode `json:"k8sNodes"`
}

var telemetryEnvMetrics *hostMetrics
var telemetryEnvMetricsOnce sync.Once

func getLocalIPs() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip != nil && !ip.IsLoopback() {
				ips = append(ips, ip)
			}
		}
	}
	return ips, nil
}

func getK8sNodes() ([]corev1.Node, error) {
	var err error
	var config *rest.Config

	if _, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST"); exist {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		home := homedir.HomeDir()
		configPath := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, err
		}
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	nodes, err := k8sClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

func getHostMetrics() *hostMetrics {
	telemetryEnvMetrics = &hostMetrics{
		IPHashes: []string{},
		K8sNodes: []models.K8sNode{},
	}
	telemetryEnvMetricsOnce.Do(func() {
		ips, err := getLocalIPs()
		if err == nil {
			for _, ip := range ips {
				md5Hash := md5.Sum([]byte(ip.String()))
				telemetryEnvMetrics.IPHashes = append(telemetryEnvMetrics.IPHashes, hex.EncodeToString(md5Hash[:]))
			}
		}
		k8sNodes, err := getK8sNodes()
		if err == nil {
			for _, node := range k8sNodes {
				telemetryEnvMetrics.K8sNodes = append(telemetryEnvMetrics.K8sNodes, models.K8sNode{
					KernelVersion:           node.Status.NodeInfo.KernelVersion,
					OsImage:                 node.Status.NodeInfo.OSImage,
					ContainerRuntimeVersion: node.Status.NodeInfo.ContainerRuntimeVersion,
					KubeletVersion:          node.Status.NodeInfo.KubeletVersion,
					KubeProxyVersion:        node.Status.NodeInfo.KubeProxyVersion,
					OperatingSystem:         node.Status.NodeInfo.OperatingSystem,
					Architecture:            node.Status.NodeInfo.Architecture,
				})
			}
		}
	})
	return telemetryEnvMetrics
}
