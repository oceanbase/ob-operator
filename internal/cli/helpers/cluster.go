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

package helpers

import (
	"fmt"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultNamespace    = "default"
	DefaultStorageClass = "local-path"
	DefaultImage        = "oceanbasedev/oceanbase-cn:4.1.0.1-test"
	GB                  = 1 << 30
)

func NewOBCluster(name string, zoneNum int, serverNum int) *v1alpha1.OBCluster {
	observerResource := &apitypes.ResourceSpec{
		Cpu:    *resource.NewQuantity(2, resource.DecimalSI),
		Memory: *resource.NewQuantity(10*GB, resource.BinarySI),
	}
	observerStorage := &apitypes.OceanbaseStorageSpec{
		DataStorage: &apitypes.StorageSpec{
			StorageClass: DefaultStorageClass,
			Size:         *resource.NewQuantity(50*GB, resource.BinarySI),
		},
		RedoLogStorage: &apitypes.StorageSpec{
			StorageClass: DefaultStorageClass,
			Size:         *resource.NewQuantity(50*GB, resource.BinarySI),
		},
		LogStorage: &apitypes.StorageSpec{
			StorageClass: DefaultStorageClass,
			Size:         *resource.NewQuantity(10*GB, resource.BinarySI),
		},
	}

	observerTemplate := &v1alpha1.OBServerTemplate{
		Image:    DefaultImage,
		Resource: observerResource,
		Storage:  observerStorage,
	}

	topology := make([]apitypes.OBZoneTopology, zoneNum)
	for i := 0; i < zoneNum; i++ {
		zoneTopology := apitypes.OBZoneTopology{
			Zone:    fmt.Sprintf("zone%d", i),
			Replica: serverNum,
		}
		topology[i] = zoneTopology
	}

	userSecrets := &apitypes.OBUserSecrets{
		Root:     "root-secret",
		ProxyRO:  "proxyro-secret",
		Monitor:  "monitor-secret",
		Operator: "operator-secret",
	}

	obcluster := &v1alpha1.OBCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OBCluster",
			APIVersion: "oceanbase.oceanbase.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: DefaultNamespace,
			Labels: map[string]string{
				"oceanbase.oceanbase.com/demo": name,
			},
		},
		Spec: v1alpha1.OBClusterSpec{
			ClusterName:      name,
			ClusterId:        1,
			OBServerTemplate: observerTemplate,
			Topology:         topology,
			UserSecrets:      userSecrets,
		},
	}
	return obcluster
}
