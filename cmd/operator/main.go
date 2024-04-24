/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"os"

	//+kubebuilder:scaffold:imports

	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/controller"
	"github.com/oceanbase/ob-operator/internal/controller/config"
	"github.com/oceanbase/ob-operator/internal/telemetry"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case -2:
		enc.AppendString("TRACE")
	default:
		enc.AppendString(level.CapitalString())
	}
}

func main() {
	var namespace string
	var managerNamespace string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var logVerbosity int
	pflag.StringVar(&namespace, "namespace", "", "The namespace to run oceanbase, default value is empty means all.")
	pflag.StringVar(&managerNamespace, "manager-namespace", "oceanbase-system", "The namespace to run manager tools.")
	pflag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	pflag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	pflag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	pflag.IntVar(&logVerbosity, "log-verbosity", 0, "Log verbosity level, 0 is info, 1 is debug, 2 is trace")
	pflag.Parse()

	opts := zap.Options{
		Development: logVerbosity > 0,
		Level:       zapcore.Level(-logVerbosity),
		EncoderConfigOptions: []zap.EncoderConfigOption{
			func(ec *zapcore.EncoderConfig) {
				ec.EncodeLevel = customLevelEncoder
			},
		},
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		Namespace:               namespace,
		MetricsBindAddress:      metricsAddr,
		Port:                    9443,
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: managerNamespace,
		LeaderElectionID:        "operator.oceanbase.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.OBClusterReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBClusterControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBCluster")
		os.Exit(1)
	}
	if err = (&controller.OBZoneReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBZoneControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBZone")
		os.Exit(1)
	}
	if err = (&controller.OBServerReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBServerControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBServer")
		os.Exit(1)
	}
	if err = (&controller.OBParameterReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBParameterControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBParameter")
		os.Exit(1)
	}
	if err = (&controller.OBTenantReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBTenantControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBTenant")
		os.Exit(1)
	}
	if err = (&controller.OBTenantBackupReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: telemetry.NewRecorder(context.Background(), mgr.GetEventRecorderFor(config.OBTenantBackupControllerName)),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBTenantBackup")
		os.Exit(1)
	}
	if err = (&controller.OBTenantRestoreReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBTenantRestoreControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBTenantRestore")
		os.Exit(1)
	}
	if err = (&controller.OBTenantBackupPolicyReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBTenantBackupPolicyControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBTenantBackupPolicy")
		os.Exit(1)
	}
	if err = (&controller.OBTenantOperationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(config.OBTenantOperationControllerName),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBTenantOperation")
		os.Exit(1)
	}
	if err = (controller.NewOBResourceRescueReconciler(mgr)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OBResourceRescue")
		os.Exit(1)
	}
	if os.Getenv("DISABLE_WEBHOOKS") != "true" {
		if err = (&v1alpha1.OBTenantBackupPolicy{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create webhook", "webhook", "OBTenantBackupPolicy")
			os.Exit(1)
		}
		if err = (&v1alpha1.OBTenant{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create webhook", "webhook", "OBTenant")
			os.Exit(1)
		}
		if err = (&v1alpha1.OBTenantOperation{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create webhook", "webhook", "OBTenantOperation")
			os.Exit(1)
		}
		if err = (&v1alpha1.OBCluster{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create webhook", "webhook", "OBCluster")
			os.Exit(1)
		}
		if err = (&v1alpha1.OBResourceRescue{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create webhook", "webhook", "OBResourceRescue")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	rcd := telemetry.NewRecorder(context.Background(), mgr.GetEventRecorderFor("ob-operator"))
	rcd.GenerateTelemetryRecord(nil, telemetry.ObjectTypeOperator, "Start", "", "Start ob-operator", nil)

	setupLog.WithValues(
		"namespace", namespace,
		"manager-namespace", managerNamespace,
		"metrics-bind-address", metricsAddr,
		"health-probe-bind-address", probeAddr,
		"leader-elect", enableLeaderElection,
		"log-verbosity", logVerbosity,
	).Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Failed to start manager")
		os.Exit(1)
	}
}
