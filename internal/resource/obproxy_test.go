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

package resource_test

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	proxystatus "github.com/oceanbase/ob-operator/internal/const/status/obproxy"
	observerstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/internal/resource/obproxy"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	strategy "github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func newNopRecorder() telemetry.Recorder {
	fakeRecorder := record.NewFakeRecorder(10)
	return telemetry.NewRecorder(context.Background(), fakeRecorder)
}

var _ = Describe("OBProxy Manager", Label("obproxy"), func() {

	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	Context("Test GetTaskFlow", func() {
		It("should return create flow for New status", func() {
			obcluster := &v1alpha1.OBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.OBClusterSpec{
					ClusterName: "test-cluster",
				},
				Status: v1alpha1.OBClusterStatus{
					Status: clusterstatus.Running,
				},
			}
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Spec: v1alpha1.OBProxySpec{
					OBCluster: v1alpha1.OBClusterReference{
						Name:      "test-cluster",
						Namespace: "default",
					},
				},
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.New,
				},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(obcluster, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx:     ctx,
				OBProxy: obproxyCR,
				Client:  fakeClient,
				Logger:  &logger,
			}
			flow, err := m.GetTaskFlow()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(flow).ShouldNot(BeNil())
			Expect(string(flow.OperationContext.Name)).Should(Equal("create obproxy"))
		})

		It("should return update flow for Updating status", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Spec: v1alpha1.OBProxySpec{},
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Updating,
				},
			}

			m := &obproxy.OBProxyManager{
				Ctx:     ctx,
				OBProxy: obproxyCR,
				Logger:  &logger,
			}
			flow, err := m.GetTaskFlow()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(flow).ShouldNot(BeNil())
			Expect(string(flow.OperationContext.Name)).Should(Equal("update obproxy"))
		})

		It("should return scale flow for Scaling status", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Scaling,
				},
			}

			m := &obproxy.OBProxyManager{
				Ctx:     ctx,
				OBProxy: obproxyCR,
				Logger:  &logger,
			}
			flow, err := m.GetTaskFlow()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(flow).ShouldNot(BeNil())
			Expect(string(flow.OperationContext.Name)).Should(Equal("scale obproxy"))
		})

		It("should return delete flow for Deleting status", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Deleting,
				},
			}

			m := &obproxy.OBProxyManager{
				Ctx:     ctx,
				OBProxy: obproxyCR,
				Logger:  &logger,
			}
			flow, err := m.GetTaskFlow()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(flow).ShouldNot(BeNil())
			Expect(string(flow.OperationContext.Name)).Should(Equal("delete obproxy"))
		})

		It("should return existing flow when OperationContext is set", func() {
			existingContext := &tasktypes.OperationContext{
				Name:  "existing operation",
				Idx:   2,
				Tasks: []tasktypes.TaskName{"CheckOBClusterReady", "CreateProxySysSecret", "CopyProxyROSecret"},
			}
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Status: v1alpha1.OBProxyStatus{
					Status:           proxystatus.New,
					OperationContext: existingContext,
				},
			}

			m := &obproxy.OBProxyManager{
				Ctx:     ctx,
				OBProxy: obproxyCR,
				Logger:  &logger,
			}
			flow, err := m.GetTaskFlow()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(flow).ShouldNot(BeNil())
			Expect(string(flow.OperationContext.Name)).Should(Equal("existing operation"))
			Expect(flow.OperationContext.Idx).Should(Equal(2))
		})

		It("should return nil for Running status", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Running,
				},
			}

			m := &obproxy.OBProxyManager{
				Ctx:     ctx,
				OBProxy: obproxyCR,
				Logger:  &logger,
			}
			flow, err := m.GetTaskFlow()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(flow).Should(BeNil())
		})
	})

	Context("Test InitStatus", func() {
		It("should initialize status correctly", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Spec: v1alpha1.OBProxySpec{
					Image:    "oceanbase/obproxy:latest",
					Replicas: 3,
				},
			}

			m := &obproxy.OBProxyManager{
				OBProxy:  obproxyCR,
				Logger:   &logger,
				Recorder: newNopRecorder(),
			}
			m.InitStatus()

			Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.New))
			Expect(obproxyCR.Status.Image).Should(Equal("oceanbase/obproxy:latest"))
			Expect(obproxyCR.Status.Replicas).Should(Equal(int32(3)))
		})
	})

	Context("Test FinishTask", func() {
		It("should set status to target status and clear operation context", func() {
			obproxyCR := &v1alpha1.OBProxy{
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Updating,
					OperationContext: &tasktypes.OperationContext{
						Name:         "update obproxy",
						TargetStatus: proxystatus.Running,
						Idx:          1,
						TaskStatus:   "running",
						Task:         "UpdateOBProxyConfigMap",
					},
				},
			}

			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}
			m.FinishTask()

			Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.Running))
			Expect(obproxyCR.Status.OperationContext).Should(BeNil())
		})
	})

	Context("Test ClearTaskInfo", func() {
		It("should clear task info and set status to running", func() {
			obproxyCR := &v1alpha1.OBProxy{
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Failed,
					OperationContext: &tasktypes.OperationContext{
						Name: "failed operation",
					},
				},
			}

			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}
			m.ClearTaskInfo()

			Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.Running))
			Expect(obproxyCR.Status.OperationContext).Should(BeNil())
		})
	})

	Context("Test GetStatus", func() {
		It("should return current status", func() {
			obproxyCR := &v1alpha1.OBProxy{
				Status: v1alpha1.OBProxyStatus{
					Status: proxystatus.Running,
				},
			}

			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}
			Expect(m.GetStatus()).Should(Equal(proxystatus.Running))
		})
	})

	Context("Test GetMeta", func() {
		It("should return object meta", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "test-ns",
				},
			}

			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}
			meta := m.GetMeta()
			Expect(meta.GetName()).Should(Equal("test-obproxy"))
			Expect(meta.GetNamespace()).Should(Equal("test-ns"))
		})
	})

	Context("Test SetOperationContext", func() {
		It("should set operation context", func() {
			obproxyCR := &v1alpha1.OBProxy{}
			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}

			opCtx := &tasktypes.OperationContext{
				Name:  "test-operation",
				Idx:   0,
				Tasks: []tasktypes.TaskName{"Task1", "Task2"},
			}
			m.SetOperationContext(opCtx)

			Expect(obproxyCR.Status.OperationContext).ShouldNot(BeNil())
			Expect(string(obproxyCR.Status.OperationContext.Name)).Should(Equal("test-operation"))
		})
	})

	Context("Test GetTaskFunc", func() {
		It("should return task function for valid task name", func() {
			obproxyCR := &v1alpha1.OBProxy{}
			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}

			taskFunc, err := m.GetTaskFunc(tasktypes.TaskName("create proxy sys secret"))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(taskFunc).ShouldNot(BeNil())
		})

		It("should return error for invalid task name", func() {
			obproxyCR := &v1alpha1.OBProxy{}
			m := &obproxy.OBProxyManager{
				OBProxy: obproxyCR,
			}

			_, err := m.GetTaskFunc(tasktypes.TaskName("InvalidTask"))
			Expect(err).Should(HaveOccurred())
		})
	})
})

var _ = Describe("OBProxy cross-namespace secret copy", Label("obproxy"), func() {

	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	Context("Test CopyProxyROSecret across namespaces", func() {
		It("should copy secret from cluster namespace to obproxy namespace", func() {
			obcluster := &v1alpha1.OBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "cluster-ns",
				},
				Spec: v1alpha1.OBClusterSpec{
					ClusterName: "test-cluster",
					UserSecrets: &apitypes.OBUserSecrets{
						ProxyRO: "proxyro-secret",
					},
				},
				Status: v1alpha1.OBClusterStatus{
					Status: clusterstatus.Running,
				},
			}

			sourceSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "proxyro-secret",
					Namespace: "cluster-ns",
				},
				Data: map[string][]byte{
					"password": []byte("proxyro-password"),
				},
			}

			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "obproxy-ns",
					UID:       types.UID("test-uid"),
				},
				Spec: v1alpha1.OBProxySpec{
					OBCluster: v1alpha1.OBClusterReference{
						Name:      "test-cluster",
						Namespace: "cluster-ns",
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(obcluster, sourceSecret, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx:      ctx,
				OBProxy:  obproxyCR,
				Client:   fakeClient,
				Recorder: newNopRecorder(),
				Logger:   &logger,
			}

			err := obproxy.CopyProxyROSecret(m)
			Expect(err).ShouldNot(HaveOccurred())

			copiedSecret := &corev1.Secret{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Namespace: "obproxy-ns",
				Name:      "sec-ro-test-obproxy",
			}, copiedSecret)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(copiedSecret.Data["password"]).Should(Equal([]byte("proxyro-password")))
		})

		It("should use obproxy namespace when cluster namespace is empty", func() {
			obcluster := &v1alpha1.OBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.OBClusterSpec{
					ClusterName: "test-cluster",
					UserSecrets: &apitypes.OBUserSecrets{
						ProxyRO: "proxyro-secret",
					},
				},
				Status: v1alpha1.OBClusterStatus{
					Status: clusterstatus.Running,
				},
			}

			sourceSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "proxyro-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"password": []byte("proxyro-password"),
				},
			}

			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
					UID:       types.UID("test-uid"),
				},
				Spec: v1alpha1.OBProxySpec{
					OBCluster: v1alpha1.OBClusterReference{
						Name:      "test-cluster",
						Namespace: "",
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(obcluster, sourceSecret, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx:      ctx,
				OBProxy:  obproxyCR,
				Client:   fakeClient,
				Recorder: newNopRecorder(),
				Logger:   &logger,
			}

			err := obproxy.CopyProxyROSecret(m)
			Expect(err).ShouldNot(HaveOccurred())

			copiedSecret := &corev1.Secret{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Namespace: "default",
				Name:      "sec-ro-test-obproxy",
			}, copiedSecret)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should return error when obcluster does not have proxyRO secret configured", func() {
			obcluster := &v1alpha1.OBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "cluster-ns",
				},
				Spec: v1alpha1.OBClusterSpec{
					ClusterName: "test-cluster",
					UserSecrets: &apitypes.OBUserSecrets{
						ProxyRO: "",
					},
				},
				Status: v1alpha1.OBClusterStatus{
					Status: clusterstatus.Running,
				},
			}

			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "obproxy-ns",
				},
				Spec: v1alpha1.OBProxySpec{
					OBCluster: v1alpha1.OBClusterReference{
						Name:      "test-cluster",
						Namespace: "cluster-ns",
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(obcluster, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx:      ctx,
				OBProxy:  obproxyCR,
				Client:   fakeClient,
				Recorder: newNopRecorder(),
				Logger:   &logger,
			}

			err := obproxy.CopyProxyROSecret(m)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("does not have proxyRO secret configured"))
		})
	})
})

var _ = Describe("OBProxy RS_LIST calculation", Label("obproxy"), func() {

	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	Context("Test RS_LIST with multiple observers", func() {
		It("should include all running observers in RS_LIST", func() {
			obcluster := &v1alpha1.OBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.OBClusterSpec{
					ClusterName: "test-cluster",
				},
				Status: v1alpha1.OBClusterStatus{
					Status: clusterstatus.Running,
				},
			}

			observers := make([]runtime.Object, 3)
			for i := 0; i < 3; i++ {
				observers[i] = &v1alpha1.OBServer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("observer-%d", i+1),
						Namespace: "default",
						Labels: map[string]string{
							oceanbaseconst.LabelRefOBCluster: "test-cluster",
						},
					},
					Status: v1alpha1.OBServerStatus{
						Status: observerstatus.Running,
						PodIp:  fmt.Sprintf("192.168.1.%d", i+1),
					},
				}
			}

			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Spec: v1alpha1.OBProxySpec{
					OBCluster: v1alpha1.OBClusterReference{
						Name:      "test-cluster",
						Namespace: "default",
					},
				},
			}

			objs := append([]runtime.Object{obcluster, obproxyCR}, observers...)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()

			m := &obproxy.OBProxyManager{
				Ctx:      ctx,
				OBProxy:  obproxyCR,
				Client:   fakeClient,
				Recorder: newNopRecorder(),
				Logger:   &logger,
			}

			err := obproxy.CreateOBProxyConfigMap(m)
			_ = err
		})

		It("should only include running observers in RS_LIST", func() {
			obcluster := &v1alpha1.OBCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.OBClusterSpec{
					ClusterName: "test-cluster",
				},
				Status: v1alpha1.OBClusterStatus{
					Status: clusterstatus.Running,
				},
			}

			runningObserver1 := &v1alpha1.OBServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "observer-1",
					Namespace: "default",
					Labels: map[string]string{
						oceanbaseconst.LabelRefOBCluster: "test-cluster",
					},
				},
				Status: v1alpha1.OBServerStatus{
					Status: observerstatus.Running,
					PodIp:  "192.168.1.1",
				},
			}

			runningObserver2 := &v1alpha1.OBServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "observer-2",
					Namespace: "default",
					Labels: map[string]string{
						oceanbaseconst.LabelRefOBCluster: "test-cluster",
					},
				},
				Status: v1alpha1.OBServerStatus{
					Status: observerstatus.Running,
					PodIp:  "192.168.1.2",
				},
			}

			notRunningObserver := &v1alpha1.OBServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "observer-3",
					Namespace: "default",
					Labels: map[string]string{
						oceanbaseconst.LabelRefOBCluster: "test-cluster",
					},
				},
				Status: v1alpha1.OBServerStatus{
					Status: "NotRunning",
					PodIp:  "192.168.1.3",
				},
			}

			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
				},
				Spec: v1alpha1.OBProxySpec{
					OBCluster: v1alpha1.OBClusterReference{
						Name:      "test-cluster",
						Namespace: "default",
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(obcluster, obproxyCR, runningObserver1, runningObserver2, notRunningObserver).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx:      ctx,
				OBProxy:  obproxyCR,
				Client:   fakeClient,
				Recorder: newNopRecorder(),
				Logger:   &logger,
			}

			err := obproxy.CreateOBProxyConfigMap(m)
			_ = err
		})
	})
})

var _ = Describe("OBProxy GetTaskFlow edge cases", Label("obproxy"), func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	It("should return nil when OBCluster is not found for New status", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				OBCluster: v1alpha1.OBClusterReference{
					Name:      "missing-cluster",
					Namespace: "default",
				},
			},
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.New,
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Client:  fakeClient,
			Logger:  &logger,
		}
		flow, err := m.GetTaskFlow()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(flow).Should(BeNil())
	})

	It("should return nil when OBCluster is not Running for New status", func() {
		obcluster := &v1alpha1.OBCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
			},
			Status: v1alpha1.OBClusterStatus{
				Status: "Bootstrapping",
			},
		}
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				OBCluster: v1alpha1.OBClusterReference{
					Name:      "test-cluster",
					Namespace: "default",
				},
			},
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.New,
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obcluster, obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Client:  fakeClient,
			Logger:  &logger,
		}
		flow, err := m.GetTaskFlow()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(flow).Should(BeNil())
	})

	It("should return nil for unknown status", func() {
		obproxyCR := &v1alpha1.OBProxy{
			Status: v1alpha1.OBProxyStatus{
				Status: "SomeUnknownStatus",
			},
		}
		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Logger:  &logger,
		}
		flow, err := m.GetTaskFlow()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(flow).Should(BeNil())
	})
})

var _ = Describe("OBProxy HandleFailure", Label("obproxy"), func() {
	It("StartOver with different status clears context and sets new status", func() {
		obproxyCR := &v1alpha1.OBProxy{
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.Failed,
				OperationContext: &tasktypes.OperationContext{
					Name:       "create obproxy",
					Idx:        2,
					TaskStatus: taskstatus.Failed,
					OnFailure: tasktypes.FailureRule{
						Strategy:      strategy.StartOver,
						NextTryStatus: proxystatus.New,
					},
				},
			},
		}
		m := &obproxy.OBProxyManager{OBProxy: obproxyCR}
		m.HandleFailure()

		Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.New))
		Expect(obproxyCR.Status.OperationContext).Should(BeNil())
	})

	It("StartOver with same status resets task progress without clearing context", func() {
		obproxyCR := &v1alpha1.OBProxy{
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.New,
				OperationContext: &tasktypes.OperationContext{
					Name:       "create obproxy",
					Idx:        3,
					TaskStatus: taskstatus.Failed,
					Task:       "CreateOBProxyDeployment",
					TaskId:     "some-task-id",
					OnFailure: tasktypes.FailureRule{
						Strategy:      strategy.StartOver,
						NextTryStatus: proxystatus.New,
					},
				},
			},
		}
		m := &obproxy.OBProxyManager{OBProxy: obproxyCR}
		m.HandleFailure()

		Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.New))
		Expect(obproxyCR.Status.OperationContext).ShouldNot(BeNil())
		Expect(obproxyCR.Status.OperationContext.Idx).Should(Equal(0))
		Expect(string(obproxyCR.Status.OperationContext.TaskStatus)).Should(BeEmpty())
		Expect(string(obproxyCR.Status.OperationContext.TaskId)).Should(BeEmpty())
		Expect(string(obproxyCR.Status.OperationContext.Task)).Should(BeEmpty())
	})

	It("RetryFromCurrent sets task status to Pending", func() {
		obproxyCR := &v1alpha1.OBProxy{
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.Updating,
				OperationContext: &tasktypes.OperationContext{
					Name:       "update obproxy",
					Idx:        1,
					TaskStatus: taskstatus.Failed,
					OnFailure: tasktypes.FailureRule{
						Strategy: strategy.RetryFromCurrent,
					},
				},
			},
		}
		m := &obproxy.OBProxyManager{OBProxy: obproxyCR}
		m.HandleFailure()

		Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.Updating))
		Expect(obproxyCR.Status.OperationContext).ShouldNot(BeNil())
		Expect(string(obproxyCR.Status.OperationContext.TaskStatus)).Should(Equal(taskstatus.Pending))
	})

	It("Pause leaves status and context unchanged", func() {
		obproxyCR := &v1alpha1.OBProxy{
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.Failed,
				OperationContext: &tasktypes.OperationContext{
					Name:       "some operation",
					Idx:        2,
					TaskStatus: taskstatus.Failed,
					OnFailure: tasktypes.FailureRule{
						Strategy: strategy.Pause,
					},
				},
			},
		}
		m := &obproxy.OBProxyManager{OBProxy: obproxyCR}
		m.HandleFailure()

		Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.Failed))
		Expect(obproxyCR.Status.OperationContext).ShouldNot(BeNil())
		Expect(obproxyCR.Status.OperationContext.Idx).Should(Equal(2))
		Expect(string(obproxyCR.Status.OperationContext.TaskStatus)).Should(Equal(taskstatus.Failed))
	})
})

var _ = Describe("OBProxy ArchiveResource", Label("obproxy"), func() {
	It("sets status to Failed and clears OperationContext", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.Updating,
				OperationContext: &tasktypes.OperationContext{
					Name: "some operation",
				},
			},
		}
		logger := logr.Discard()
		m := &obproxy.OBProxyManager{
			OBProxy:  obproxyCR,
			Logger:   &logger,
			Recorder: newNopRecorder(),
		}
		m.ArchiveResource()

		Expect(obproxyCR.Status.Status).Should(Equal(proxystatus.Failed))
		Expect(obproxyCR.Status.OperationContext).Should(BeNil())
	})
})

var _ = Describe("OBProxy CheckAndUpdateFinalizers", Label("obproxy"), func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
	})

	It("clears finalizers and updates when status is FinalizerFinished", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-obproxy",
				Namespace:  "default",
				Finalizers: []string{"oceanbase.com/finalizer"},
			},
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.FinalizerFinished,
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Client:  fakeClient,
		}
		Expect(m.CheckAndUpdateFinalizers()).Should(Succeed())
		Expect(obproxyCR.ObjectMeta.Finalizers).Should(BeEmpty())
	})

	It("does nothing for non-FinalizerFinished status", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-obproxy",
				Namespace:  "default",
				Finalizers: []string{"oceanbase.com/finalizer"},
			},
			Status: v1alpha1.OBProxyStatus{
				Status: proxystatus.Running,
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Client:  fakeClient,
		}
		Expect(m.CheckAndUpdateFinalizers()).Should(Succeed())
		Expect(obproxyCR.ObjectMeta.Finalizers).Should(HaveLen(1))
	})
})

var _ = Describe("OBProxy CreateOBProxyConfigMap task", Label("obproxy"), func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	It("creates ConfigMap with uppercased ODP_ parameter keys", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				Parameters: []apitypes.Parameter{
					{Name: "max_connections", Value: "100"},
					{Name: "enable_transaction_internal_routing", Value: "true"},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:      ctx,
			OBProxy:  obproxyCR,
			Client:   fakeClient,
			Recorder: newNopRecorder(),
			Logger:   &logger,
		}
		Expect(obproxy.CreateOBProxyConfigMap(m)).Should(Succeed())

		cm := &corev1.ConfigMap{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "cm-test-obproxy",
		}, cm)).Should(Succeed())
		Expect(cm.Data["ODP_MAX_CONNECTIONS"]).Should(Equal("100"))
		Expect(cm.Data["ODP_ENABLE_TRANSACTION_INTERNAL_ROUTING"]).Should(Equal("true"))
	})

	It("skips creation and preserves existing ConfigMap", func() {
		existingCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm-test-obproxy",
				Namespace: "default",
			},
			Data: map[string]string{"ODP_MAX_CONNECTIONS": "50"},
		}
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				Parameters: []apitypes.Parameter{
					{Name: "max_connections", Value: "100"},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(existingCM, obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:      ctx,
			OBProxy:  obproxyCR,
			Client:   fakeClient,
			Recorder: newNopRecorder(),
			Logger:   &logger,
		}
		Expect(obproxy.CreateOBProxyConfigMap(m)).Should(Succeed())

		cm := &corev1.ConfigMap{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "cm-test-obproxy",
		}, cm)).Should(Succeed())
		Expect(cm.Data["ODP_MAX_CONNECTIONS"]).Should(Equal("50"))
	})
})

var _ = Describe("OBProxy CreateOBProxyService task", Label("obproxy"), func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	It("creates Service with correct ports and default ClusterIP type", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:      ctx,
			OBProxy:  obproxyCR,
			Client:   fakeClient,
			Recorder: newNopRecorder(),
			Logger:   &logger,
		}
		Expect(obproxy.CreateOBProxyService(m)).Should(Succeed())

		svc := &corev1.Service{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "svc-test-obproxy",
		}, svc)).Should(Succeed())
		Expect(svc.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))

		portMap := map[string]int32{}
		for _, p := range svc.Spec.Ports {
			portMap[p.Name] = p.Port
		}
		Expect(portMap["sql"]).Should(Equal(int32(2883)))
		Expect(portMap["prometheus"]).Should(Equal(int32(2884)))
	})

	It("creates Service with LoadBalancer type when specified", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				ServiceType: "LoadBalancer",
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:      ctx,
			OBProxy:  obproxyCR,
			Client:   fakeClient,
			Recorder: newNopRecorder(),
			Logger:   &logger,
		}
		Expect(obproxy.CreateOBProxyService(m)).Should(Succeed())

		svc := &corev1.Service{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "svc-test-obproxy",
		}, svc)).Should(Succeed())
		Expect(svc.Spec.Type).Should(Equal(corev1.ServiceTypeLoadBalancer))
	})

	It("skips creation and preserves existing Service", func() {
		existingSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc-test-obproxy",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeNodePort},
		}
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(existingSvc, obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:      ctx,
			OBProxy:  obproxyCR,
			Client:   fakeClient,
			Recorder: newNopRecorder(),
			Logger:   &logger,
		}
		Expect(obproxy.CreateOBProxyService(m)).Should(Succeed())

		svc := &corev1.Service{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "svc-test-obproxy",
		}, svc)).Should(Succeed())
		Expect(svc.Spec.Type).Should(Equal(corev1.ServiceTypeNodePort))
	})
})

var _ = Describe("OBProxy UpdateOBProxyConfigMap task", Label("obproxy"), func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	It("updates existing ConfigMap with new parameters and removes stale keys", func() {
		existingCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm-test-obproxy",
				Namespace: "default",
				Labels: map[string]string{
					obproxy.LabelOBProxyInstance: "test-obproxy",
				},
			},
			Data: map[string]string{"ODP_OLD_PARAM": "old-value"},
		}
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				Parameters: []apitypes.Parameter{
					{Name: "max_connections", Value: "200"},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(existingCM, obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Client:  fakeClient,
			Logger:  &logger,
		}
		Expect(obproxy.UpdateOBProxyConfigMap(m)).Should(Succeed())

		cm := &corev1.ConfigMap{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "cm-test-obproxy",
		}, cm)).Should(Succeed())
		Expect(cm.Data["ODP_MAX_CONNECTIONS"]).Should(Equal("200"))
		Expect(cm.Data).ShouldNot(HaveKey("ODP_OLD_PARAM"))
	})

	It("creates ConfigMap when it does not exist", func() {
		obproxyCR := &v1alpha1.OBProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-obproxy",
				Namespace: "default",
			},
			Spec: v1alpha1.OBProxySpec{
				Parameters: []apitypes.Parameter{
					{Name: "max_connections", Value: "50"},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(obproxyCR).
			Build()

		m := &obproxy.OBProxyManager{
			Ctx:     ctx,
			OBProxy: obproxyCR,
			Client:  fakeClient,
			Logger:  &logger,
		}
		Expect(obproxy.UpdateOBProxyConfigMap(m)).Should(Succeed())

		cm := &corev1.ConfigMap{}
		Expect(fakeClient.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "cm-test-obproxy",
		}, cm)).Should(Succeed())
		Expect(cm.Data["ODP_MAX_CONNECTIONS"]).Should(Equal("50"))
	})
})

var _ = Describe("OBProxy delete tasks", Label("obproxy"), func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		logger logr.Logger
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
		Expect(corev1.AddToScheme(scheme)).Should(Succeed())
		Expect(appsv1.AddToScheme(scheme)).Should(Succeed())
		ctx = context.Background()
		logger = logr.Discard()
	})

	Context("DeleteOBProxyDeployment", func() {
		It("deletes an existing deployment", func() {
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-obproxy",
					Namespace: "default",
					Labels: map[string]string{
						obproxy.LabelOBProxyInstance: "test-obproxy",
					},
				},
			}
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(deploy, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxyDeployment(m)).Should(Succeed())

			list := &appsv1.DeploymentList{}
			Expect(fakeClient.List(ctx, list)).Should(Succeed())
			Expect(list.Items).Should(BeEmpty())
		})

		It("succeeds when deployment does not exist", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).WithRuntimeObjects(obproxyCR).Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxyDeployment(m)).Should(Succeed())
		})
	})

	Context("DeleteOBProxyService", func() {
		It("deletes an existing service", func() {
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-test-obproxy",
					Namespace: "default",
					Labels: map[string]string{
						obproxy.LabelOBProxyInstance: "test-obproxy",
					},
				},
			}
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(svc, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxyService(m)).Should(Succeed())

			list := &corev1.ServiceList{}
			Expect(fakeClient.List(ctx, list)).Should(Succeed())
			Expect(list.Items).Should(BeEmpty())
		})

		It("succeeds when service does not exist", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).WithRuntimeObjects(obproxyCR).Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxyService(m)).Should(Succeed())
		})
	})

	Context("DeleteOBProxyConfigMap", func() {
		It("deletes an existing configmap", func() {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cm-test-obproxy",
					Namespace: "default",
					Labels: map[string]string{
						obproxy.LabelOBProxyInstance: "test-obproxy",
					},
				},
			}
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(cm, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxyConfigMap(m)).Should(Succeed())

			list := &corev1.ConfigMapList{}
			Expect(fakeClient.List(ctx, list)).Should(Succeed())
			Expect(list.Items).Should(BeEmpty())
		})

		It("succeeds when configmap does not exist", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).WithRuntimeObjects(obproxyCR).Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxyConfigMap(m)).Should(Succeed())
		})
	})

	Context("DeleteOBProxySecrets", func() {
		It("deletes the proxyRO secret", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sec-ro-test-obproxy",
					Namespace: "default",
				},
			}
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(secret, obproxyCR).
				Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxySecrets(m)).Should(Succeed())

			deleted := &corev1.Secret{}
			err := fakeClient.Get(ctx, types.NamespacedName{
				Namespace: "default", Name: "sec-ro-test-obproxy",
			}, deleted)
			Expect(err).Should(HaveOccurred())
		})

		It("succeeds when proxyRO secret does not exist", func() {
			obproxyCR := &v1alpha1.OBProxy{
				ObjectMeta: metav1.ObjectMeta{Name: "test-obproxy", Namespace: "default"},
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).WithRuntimeObjects(obproxyCR).Build()

			m := &obproxy.OBProxyManager{
				Ctx: ctx, OBProxy: obproxyCR, Client: fakeClient,
				Recorder: newNopRecorder(), Logger: &logger,
			}
			Expect(obproxy.DeleteOBProxySecrets(m)).Should(Succeed())
		})
	})
})
