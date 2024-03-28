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

package oceanbase

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func GetOBClusterUsages(ctx context.Context, nn *param.K8sObjectIdentity) (*response.OBClusterResources, error) {
	obcluster, err := clients.GetOBCluster(ctx, nn.Namespace, nn.Name)
	if err != nil {
		return nil, err
	}
	clt := client.GetClient()
	serverList := &v1alpha1.OBServerList{}
	err = clients.ServerClient.List(ctx, nn.Namespace, serverList, metav1.ListOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	rootSecret, err := clt.ClientSet.CoreV1().Secrets(nn.Namespace).Get(ctx, obcluster.Spec.UserSecrets.Root, metav1.GetOptions{})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	password, ok := rootSecret.Data["password"]
	if !ok {
		return nil, httpErr.NewInternal("root password not found")
	}
	var manager *operation.OceanbaseOperationManager
	for _, observer := range serverList.Items {
		source := connector.NewOceanBaseDataSource(observer.Status.GetConnectAddr(), oceanbaseconst.SqlPort, "root", "sys", string(password), oceanbaseconst.DefaultDatabase)
		manager, err = operation.GetOceanbaseOperationManager(source)
		if err == nil {
			break
		}
	}
	if manager == nil {
		return nil, httpErr.NewInternal("no running observer is connectable")
	}
	defer manager.Close()

	parameters, err := manager.GetParameter("__min_full_resource_pool_memory", nil)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}

	essentials := &response.OBClusterResources{}
	if len(parameters) == 0 {
		essentials.MinPoolMemory = 5 << 30 // 5 Gi
	} else {
		minPoolMemory, err := resource.ParseQuantity(parameters[0].Value)
		if err != nil {
			return nil, httpErr.NewInternal(err.Error())
		}
		essentials.MinPoolMemory = minPoolMemory.Value()
	}
	gvservers, err := manager.ListGVServers()
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	serverUsages, zoneMapping := getServerUsages(gvservers)
	essentials.OBServerResources = serverUsages
	essentials.OBZoneResourceMap = zoneMapping
	return essentials, nil
}

func getServerUsages(gvservers []model.GVOBServer) ([]response.OBServerAvailableResource, map[string]*response.OBZoneAvaiableResource) {
	zoneMapping := make(map[string]*response.OBZoneAvaiableResource)
	serverUsages := make([]response.OBServerAvailableResource, 0, len(gvservers))
	for _, gvserver := range gvservers {
		zoneResource := &response.OBZoneAvaiableResource{
			ServerCount:       1,
			OBZone:            gvserver.Zone,
			AvailableCPU:      max(gvserver.CPUCapacity-gvserver.CPUAssigned, 0),
			AvailableMemory:   max(gvserver.MemCapacity-gvserver.MemAssigned, 0),
			AvailableLogDisk:  max(gvserver.LogDiskCapacity-gvserver.LogDiskAssigned, 0),
			AvailableDataDisk: max(gvserver.DataDiskCapacity-gvserver.DataDiskAllocated, 0),
		}
		serverUsage := response.OBServerAvailableResource{
			OBServerIP:             gvserver.ServerIP,
			OBZoneAvaiableResource: *zoneResource,
		}
		if _, ok := zoneMapping[gvserver.Zone]; !ok {
			zoneMapping[gvserver.Zone] = zoneResource
		} else {
			zoneMapping[gvserver.Zone].ServerCount++
			if zoneMapping[gvserver.Zone].AvailableCPU < serverUsage.AvailableCPU {
				zoneMapping[gvserver.Zone].AvailableCPU = serverUsage.AvailableCPU
			}
			if zoneMapping[gvserver.Zone].AvailableMemory < serverUsage.AvailableMemory {
				zoneMapping[gvserver.Zone].AvailableMemory = serverUsage.AvailableMemory
			}
			if zoneMapping[gvserver.Zone].AvailableLogDisk < serverUsage.AvailableLogDisk {
				zoneMapping[gvserver.Zone].AvailableLogDisk = serverUsage.AvailableLogDisk
			}
			if zoneMapping[gvserver.Zone].AvailableDataDisk < serverUsage.AvailableDataDisk {
				zoneMapping[gvserver.Zone].AvailableDataDisk = serverUsage.AvailableDataDisk
			}
		}
		serverUsages = append(serverUsages, serverUsage)
	}
	return serverUsages, zoneMapping
}

type OrderedType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func max[t OrderedType](a, b t) t {
	if a > b {
		return a
	}
	return b
}
