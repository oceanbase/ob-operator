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

package kubeclient

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewClientFromManager(mgr manager.Manager) client.Client {
	cfg := *mgr.GetConfig()

	// cfg.Burst = 10
	// cfg.QPS = float32(10)

	c, err := client.New(
		&cfg,
		client.Options{Scheme: mgr.GetScheme(), Mapper: mgr.GetRESTMapper()},
	)
	if err != nil {
		panic(err)
	}

	delegatingClient, _ := client.NewDelegatingClient(
		client.NewDelegatingClientInput{
			CacheReader: mgr.GetCache(),
			Client:      c,
		},
	)

	return delegatingClient
}
