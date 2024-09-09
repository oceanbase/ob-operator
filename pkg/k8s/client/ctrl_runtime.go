/*
Copyright (c) 2024 OceanBase
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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrlruntime "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type addToSchemeFunc func(s *runtime.Scheme) error

func GetCtrlRuntimeClient(config *rest.Config, adds ...addToSchemeFunc) (ctrlruntime.Client, error) {
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add client-go scheme")
	}
	for _, add := range adds {
		err = add(scheme)
		if err != nil {
			return nil, errors.Wrap(err, "failed to add custom scheme")
		}
	}
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http client")
	}
	restMapper, err := apiutil.NewDiscoveryRESTMapper(config, httpClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create rest mapper")
	}
	return ctrlruntime.New(config,
		ctrlruntime.Options{
			Scheme: scheme,
			Mapper: restMapper,
		})
}
