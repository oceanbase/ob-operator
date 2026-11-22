/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

// oceanseek-readiness is the read-only post-install/post-upgrade Helm hook.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oceanbase/ob-operator/internal/oceanseek/readiness"
)

func main() {
	var o readiness.Options
	var proxySeedsFile string
	flag.StringVar(&proxySeedsFile, "obproxy-rslist-file", "", "Write advertised Observer Service addresses for OBProxy bootstrap")
	flag.StringVar(&o.Namespace, "namespace", "", "Kubernetes namespace")
	flag.StringVar(&o.Release, "release", "", "Helm release name")
	flag.StringVar(&o.StorageJob, "storage-job", "", "Storage initialization Job name")
	flag.StringVar(&o.Dashboard, "dashboard", "", "Dashboard Deployment name; empty disables this check")
	flag.IntVar(&o.ObserverReplicas, "observer-replicas", 1, "Expected Observer Pods")
	flag.IntVar(&o.LogServiceReplicas, "logservice-replicas", 3, "Expected LogService Pods")
	flag.DurationVar(&o.Timeout, "timeout", 30*time.Minute, "Overall readiness timeout")
	flag.DurationVar(&o.Interval, "interval", 5*time.Second, "Readiness poll interval")
	flag.Parse()
	if proxySeedsFile != "" {
		o.StorageJob = o.Release + "-storage-init"
	}
	if err := o.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	client, err := readiness.InClusterClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if proxySeedsFile != "" {
		if err := readiness.WriteProxySeeds(ctx, client, o, proxySeedsFile); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := readiness.Wait(ctx, client, o, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Readiness check failed:", err)
		os.Exit(1)
	}
}
