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

package controller

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/pkg/controller/config"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sManager ctrl.Manager
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	t := true
	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
		testEnv = &envtest.Environment{
			UseExistingCluster: &t,
		}
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
		}
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// TODO whether to deleted
	// err = scheme.AddToScheme(scheme.Scheme)
	// Expect(err).NotTo(HaveOccurred())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	if os.Getenv("TEST_USE_EXISTING_CLUSTER") != "true" {
		// init k8s manager and setup controllers
		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
		logf.Log.Error(err, "create manager failed")
		Expect(err).NotTo(HaveOccurred())

		err = (&OBClusterReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBClusterControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBZoneReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBZoneControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBServerReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBServerControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBParameterReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBParameterControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBTenantReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBTenantControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBUnitReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBUnitControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBClusterBackupReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBClusterBackupControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBTenantBackupReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBTenantBackupControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBClusterRestoreReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBClusterRestoreControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		err = (&OBTenantRestoreReconciler{
			Client:   k8sManager.GetClient(),
			Scheme:   k8sManager.GetScheme(),
			Recorder: k8sManager.GetEventRecorderFor(config.OBTenantRestoreControllerName),
		}).SetupWithManager(k8sManager)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			err = k8sManager.Start(ctrl.SetupSignalHandler())
			Expect(err).ToNot(HaveOccurred())
		}()
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
