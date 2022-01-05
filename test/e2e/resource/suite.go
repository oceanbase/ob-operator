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

package resource

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
)

var cfg *rest.Config
var TestClient *Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"ob-operator test suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	config.DefaultReporterConfig.SlowSpecThreshold = 3600

	format.MaxDepth = 1
	format.UseStringerRepresentation = true
	format.PrintContextObjects = false
	format.TruncatedDiff = false

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	cfg, TestClient = NewClient()
	Expect(TestClient).NotTo(BeNil())

	By("bootstrapping test environment")
	var useExistingCluster bool
	useExistingCluster = true
	testEnv = &envtest.Environment{
		UseExistingCluster:    &useExistingCluster,
		Config:                cfg,
		CRDDirectoryPaths:     []string{filepath.Join("../../", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cloudv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
}, 60)

var _ = AfterSuite(func() {
	TestClient.DeleteObj(testconverter.MakeObjectFromFile("./data/statefulapp-01.yaml").(unstructured.Unstructured))
	TestClient.DeleteObj(testconverter.MakeObjectFromFile("./data/statefulapp-02.yaml").(unstructured.Unstructured))

	// TestClient.DeleteObj(testresource.MakeObjectFromFile("./data/obcluster-01.yaml").(unstructured.Unstructured))
	// TestClient.DeleteObj(testresource.MakeObjectFromFile("./data/obcluster-02.yaml").(unstructured.Unstructured))

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
