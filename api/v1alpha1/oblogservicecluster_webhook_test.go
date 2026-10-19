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

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apitypes "github.com/oceanbase/ob-operator/api/types"
)

var _ = Describe("Test OBLogServiceCluster Webhook", Label("webhook"), func() {
	It("rejects a missing object-store Secret", func() {
		cluster := newOBLogServiceCluster("logservice-missing-secret", "secret-that-does-not-exist")
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("rejects an object-store Secret without the required keys", func() {
		cluster := newOBLogServiceCluster("logservice-wrong-secret", wrongKeySecret)
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("accepts an object-store Secret with non-empty credentials", func() {
		const secretName = "logservice-object-store-secret"
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: defaultNamespace},
			Data: map[string][]byte{
				"access_id":  []byte("test-access-id"),
				"access_key": []byte("test-access-key"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

		cluster := newOBLogServiceCluster("logservice-valid-secret", secretName)
		cluster.Spec.Parameters = []apitypes.Parameter{{Name: "syslog_level", Value: "WARN"}}
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())
		By("rejecting invalid topology through the API server on update")
		for _, change := range []func(*OBLogServiceCluster){
			func(c *OBLogServiceCluster) { c.Spec.Topology[0].Replica = 0 },
			func(c *OBLogServiceCluster) { c.Spec.Topology[0].Replica = -1 },
			func(c *OBLogServiceCluster) { c.Spec.Topology = append(c.Spec.Topology, c.Spec.Topology[0]) },
			func(c *OBLogServiceCluster) { c.Spec.Topology[0].Zone = "" },
			func(c *OBLogServiceCluster) { c.Spec.Topology = nil },
			func(c *OBLogServiceCluster) {
				c.Spec.Topology[0].Resource = &apitypes.ResourceSpec{Memory: resource.MustParse("8Gi")}
			},
			func(c *OBLogServiceCluster) { c.Spec.Topology[0].NodeSelector = map[string]string{"pool": "other"} },
			func(c *OBLogServiceCluster) {
				c.Spec.Topology[0].Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}
			},
			func(c *OBLogServiceCluster) {
				c.Spec.Topology[0].Tolerations = []corev1.Toleration{{Key: "pool", Operator: corev1.TolerationOpExists}}
			},
			func(c *OBLogServiceCluster) { c.Spec.Topology[0].RpcPort = 51000 },
			func(c *OBLogServiceCluster) { c.Spec.Topology[0].HttpPort = 51001 },
			func(c *OBLogServiceCluster) { c.Spec.ServiceAccount = "other-workload" },
			func(c *OBLogServiceCluster) {
				c.Spec.Parameters = append(c.Spec.Parameters, apitypes.Parameter{Name: "new_parameter", Value: "new_value"})
			},
		} {
			invalid := cluster.DeepCopy()
			change(invalid)
			Expect(k8sClient.Update(ctx, invalid)).ShouldNot(Succeed())
		}
		By("allowing scale out, scale in, and status updates")
		cluster.Spec.Parameters[0], cluster.Spec.Parameters[1] = cluster.Spec.Parameters[1], cluster.Spec.Parameters[0]
		cluster.Spec.Topology[0].Replica = 2
		Expect(k8sClient.Update(ctx, cluster)).Should(Succeed())
		cluster.Spec.Topology[0].Replica = 1
		Expect(k8sClient.Update(ctx, cluster)).Should(Succeed())
		Expect(k8sClient.Status().Update(ctx, cluster)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, secret)).Should(Succeed())
	})
})

func newOBLogServiceCluster(name, secretName string) *OBLogServiceCluster {
	return &OBLogServiceCluster{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: defaultNamespace},
		Spec: OBLogServiceClusterSpec{
			ClusterId: 1,
			LogService: &apitypes.LogServiceTemplate{
				Image:    "oblogservice:test",
				Resource: apitypes.ResourceSpec{Memory: resource.MustParse("4Gi")},
				Storage: &apitypes.LogServiceStorageSpec{
					StoreStorage: &apitypes.StorageSpec{StorageClass: defaultStorageClass, Size: resource.MustParse("20Gi")},
					LogStorage:   &apitypes.StorageSpec{StorageClass: defaultStorageClass, Size: resource.MustParse("10Gi")},
				},
			},
			Topology: []apitypes.LogServiceZoneTopology{{Zone: "zone1", Replica: 1}},
			ObjectStoreURL: apitypes.ObjectStoreConfig{
				BucketURL: "s3://bucket?host=http://object-store:9000",
				SecretRef: corev1.LocalObjectReference{Name: secretName},
			},
		},
	}
}
