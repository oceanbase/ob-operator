/*
Copyright (c) 2021 OceanBase
Copyright 2021 The Kruise Authors.
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package kube

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	"github.com/oceanbase/ob-operator/apis"
	"github.com/oceanbase/ob-operator/pkg/kubeclient"
)

var (
	internalScheme = runtime.NewScheme()

	errKindNotFound = errors.Errorf("kind not found in group version resources")

	backOff = wait.Backoff{
		Steps:    4,
		Duration: 500 * time.Millisecond,
		Factor:   5.0,
		Jitter:   0.1,
	}
)

func init() {
	err := apis.AddToScheme(internalScheme)
	if err != nil {
		klog.Errorln(err)
	}
}

func DiscoverGVK(gvk schema.GroupVersionKind) bool {
	genericClient := kubeclient.GetGenericClient()
	if genericClient == nil {
		return true
	}
	discoveryClient := genericClient.DiscoveryClient
	startTime := time.Now()

	err := retry.OnError(
		backOff,
		func(err error) bool {
			return true
		},
		func() error {
			resourceList, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
			if err != nil {
				return err
			}
			for _, r := range resourceList.APIResources {
				if r.Kind == gvk.Kind {
					return nil
				}
			}
			return errKindNotFound
		},
	)

	if err != nil {
		if err == errKindNotFound {
			klog.Errorf("not found kind %s in group version %s, waiting time %s",
				gvk.Kind, gvk.GroupVersion().String(), time.Since(startTime))
			return false
		}
		// This might be caused by abnormal apiserver or etcd, ignore it
		klog.Errorf("failed to find resources in group version %s: %v, waiting time %s",
			gvk.GroupVersion().String(), err, time.Since(startTime))
	}
	return true
}
