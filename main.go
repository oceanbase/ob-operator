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

package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	controller "github.com/oceanbase/ob-operator/pkg/controllers"
	"github.com/oceanbase/ob-operator/pkg/util"
	// +kubebuilder:scaffold:imports
)

var (
	// exit func list
	funcList []func()
	scheme   = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cloudv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var clusterName string
	var namespace string
	var enableLeaderElection bool
	var managerNamespace string
	var syncPeriodStr string
	var probeAddr string
	var metricsAddr string
	var enablePprof bool
	var pprofAddr string

	flag.StringVar(&clusterName, "cluster-name", "cn",
		"Which cluster to run oceanbase. Defaults is cn.")
	flag.StringVar(&namespace, "namespace", "",
		"Which namespace to run oceanbase. Defaults is all namespace.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager.")
	flag.StringVar(&managerNamespace, "manager-namespace", "oceanbase-system",
		"Which namespace to run manager tools.")
	flag.StringVar(&syncPeriodStr, "sync-period", "5s",
		"How often should data be synchronized.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":9090",
		"The address the probe endpoint binds to.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.BoolVar(&enablePprof, "enable-pprof", false,
		"Enable pprof for controller manager.")
	flag.StringVar(&pprofAddr, "pprof-addr", ":9091",
		"The address the pprof binds to.")

	flag.Parse()

	klog.InitFlags(nil)
	ctrl.SetLogger(klogr.New())

	myconfig.ClusterName = clusterName

	// pprof
	if enablePprof {
		go func() {
			err := http.ListenAndServe(pprofAddr, nil)
			if err != nil {
				klog.Errorln(err, "unable to start pprof")
			}
		}()
	}

	// syncPeriod
	var syncPeriod *time.Duration
	if syncPeriodStr != "" {
		duration, err := time.ParseDuration(syncPeriodStr)
		if err != nil {
			klog.Errorln(err, "invalid sync period flag")
		} else {
			syncPeriod = &duration
		}
	}

	mgr, err := ctrl.NewManager(
		ctrl.GetConfigOrDie(),
		ctrl.Options{
			Scheme:                     scheme,
			Namespace:                  namespace,
			LeaderElection:             enableLeaderElection,
			LeaderElectionNamespace:    managerNamespace,
			LeaderElectionID:           "oceanbase-system",
			LeaderElectionResourceLock: resourcelock.ConfigMapsResourceLock,
			SyncPeriod:                 syncPeriod,
			HealthProbeBindAddress:     probeAddr,
			MetricsBindAddress:         metricsAddr,
		},
	)
	if err != nil {
		klog.Errorln(err, "unable to start manager.")
		os.Exit(1)
	}

	if err = controller.SetupWithManager(mgr); err != nil {
		klog.Errorln(err, "unable to setup controllers.")
		os.Exit(1)
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Errorln(err, "unable to set up health check.")
		os.Exit(1)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Errorln(err, "unable to set up ready check.")
		os.Exit(1)
	}

	klog.Infoln("starting manager.")
	if err = mgr.Start(util.SignalHandler(funcList)); err != nil {
		klog.Errorln(err, "problem running manager.")
		os.Exit(1)
	}
}
