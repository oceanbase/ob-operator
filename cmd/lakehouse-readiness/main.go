// lakehouse-readiness is the read-only post-install/post-upgrade Helm hook.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oceanbase/ob-operator/internal/lakehouse/readiness"
)

func main() {
	var o readiness.Options
	flag.StringVar(&o.Namespace, "namespace", "", "Kubernetes namespace")
	flag.StringVar(&o.Release, "release", "", "Helm release name")
	flag.StringVar(&o.StorageJob, "storage-job", "", "Storage initialization Job name")
	flag.StringVar(&o.Dashboard, "dashboard", "", "Dashboard Deployment name; empty disables this check")
	flag.IntVar(&o.ObserverReplicas, "observer-replicas", 1, "Expected Observer Pods")
	flag.IntVar(&o.LogServiceReplicas, "logservice-replicas", 3, "Expected LogService Pods")
	flag.DurationVar(&o.Timeout, "timeout", 30*time.Minute, "Overall readiness timeout")
	flag.DurationVar(&o.Interval, "interval", 5*time.Second, "Readiness poll interval")
	flag.Parse()
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
	if err := readiness.Wait(ctx, client, o, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Readiness check failed:", err)
		os.Exit(1)
	}
}
