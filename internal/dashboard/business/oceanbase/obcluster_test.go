/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanbase

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

func getMockedCreateClusterParam() *param.CreateOBClusterParam {
	return &param.CreateOBClusterParam{
		Namespace:    "default",
		Name:         "test-cluster",
		ClusterName:  "obcluster",
		ClusterId:    123,
		RootPassword: "test-password",
		Topology: []param.ZoneTopology{{
			Zone:     "zone1",
			Replicas: 1,
			NodeSelector: []common.KVPair{{
				Key:   "test-node-selector",
				Value: "test-node-selector-value",
			}},
			Tolerations: []common.KVPair{{
				Key:   "test-toleration",
				Value: "test-toleration-value",
			}},
			Affinities: []common.AffinitySpec{{
				KVPair: common.KVPair{
					Key:   "test-node-affinity",
					Value: "test-node-affinity-value",
				},
				Type: common.NodeAffinityType,
			}, {
				KVPair: common.KVPair{
					Key:   "test-pod-affinity",
					Value: "test-pod-affinity-value",
				},
				Type: common.PodAffinityType,
			}, {
				KVPair: common.KVPair{
					Key:   "test-pod-anti-affinity",
					Value: "test-pod-anti-affinity-value",
				},
				Type: common.PodAntiAffinityType,
			}},
		}, {
			Zone:     "zone2",
			Replicas: 1,
		}},
		OBServer: &param.OBServerSpec{
			Image: "oceanbasedev/oceanbase:test",
			Resource: common.ResourceSpec{
				Cpu:      2,
				MemoryGB: 10,
			},
			Storage: &param.OBServerStorageSpec{
				Data: common.StorageSpec{
					StorageClass: "local-path",
					SizeGB:       30,
				},
				RedoLog: common.StorageSpec{
					StorageClass: "local-path",
					SizeGB:       40,
				},
				Log: common.StorageSpec{
					StorageClass: "local-path",
					SizeGB:       20,
				},
			},
		},
		Monitor: &param.MonitorSpec{
			Image: "oceanbasedev/obagent:test",
			Resource: common.ResourceSpec{
				Cpu:      1,
				MemoryGB: 1,
			},
		},
		Parameters: []common.KVPair{{
			Key:   "__min_full_resource_pool_memory",
			Value: "2147483648",
		}},
		BackupVolume: &param.NFSVolumeSpec{
			Address: "1.2.3.4",
			Path:    "/path/to/mount",
		},
	}
}

func getExpectedCluster() *v1alpha1.OBCluster {
	return &v1alpha1.OBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-cluster",
		},
		Spec: v1alpha1.OBClusterSpec{
			ClusterName: "obcluster",
			ClusterId:   123,
			OBServerTemplate: &types.OBServerTemplate{
				Image: "oceanbasedev/oceanbase:test",
				Resource: &types.ResourceSpec{
					Cpu:    resource.MustParse("2"),
					Memory: resource.MustParse("10Gi"),
				},
				Storage: &types.OceanbaseStorageSpec{
					DataStorage: &types.StorageSpec{
						StorageClass: "local-path",
						Size:         resource.MustParse("30Gi"),
					},
					RedoLogStorage: &types.StorageSpec{
						StorageClass: "local-path",
						Size:         resource.MustParse("40Gi"),
					},
					LogStorage: &types.StorageSpec{
						StorageClass: "local-path",
						Size:         resource.MustParse("20Gi"),
					},
				},
			},
			MonitorTemplate: &types.MonitorTemplate{
				Image: "oceanbasedev/obagent:test",
				Resource: &types.ResourceSpec{
					Cpu:    resource.MustParse("1"),
					Memory: resource.MustParse("1Gi"),
				},
			},
			BackupVolume: &types.BackupVolumeSpec{
				Volume: &corev1.Volume{
					Name: "ob-backup",
					VolumeSource: corev1.VolumeSource{
						NFS: &corev1.NFSVolumeSource{
							Server: "1.2.3.4",
							Path:   "/path/to/mount",
						},
					},
				},
			},
			Parameters: []types.Parameter{{
				Name:  "__min_full_resource_pool_memory",
				Value: "2147483648",
			}},
			Topology: []types.OBZoneTopology{{
				Zone: "zone1",
				NodeSelector: map[string]string{
					"test-node-selector": "test-node-selector-value",
				},
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{{
								MatchExpressions: []corev1.NodeSelectorRequirement{{
									Key:      "test-node-affinity",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"test-node-affinity-value"},
								}},
							}},
						},
					},
					PodAffinity: &corev1.PodAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{{
									Key:      "test-pod-affinity",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"test-pod-affinity-value"},
								}},
							},
						}},
					},
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{{
									Key:      "test-pod-anti-affinity",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"test-pod-anti-affinity-value"},
								}},
							},
						}},
					},
				},
				Tolerations: []corev1.Toleration{{
					Key:      "test-toleration",
					Operator: corev1.TolerationOpEqual,
					Value:    "test-toleration-value",
					Effect:   corev1.TaintEffectNoSchedule,
				}},
				Replica: 1,
			}, {
				Zone:         "zone2",
				Replica:      1,
				NodeSelector: map[string]string{},
			}},
			UserSecrets: &types.OBUserSecrets{
				Root:     fmt.Sprintf("%s-%d-root-", "test-cluster", 123),
				ProxyRO:  fmt.Sprintf("%s-%d-proxyro-", "test-cluster", 123),
				Monitor:  fmt.Sprintf("%s-%d-monitor-", "test-cluster", 123),
				Operator: fmt.Sprintf("%s-%d-operator-", "test-cluster", 123),
			},
			ServiceAccount: "",
		},
	}
}

var _ = Describe("Test OBCluster", func() {
	It("Test creating obcluster", func() {
		param := getMockedCreateClusterParam()
		expected := getExpectedCluster()
		actual := generateOBClusterInstance(param)

		Expect(actual.Name).Should(Equal(expected.Name))
		Expect(actual.Namespace).Should(Equal(expected.Namespace))

		Expect(actual.Spec.ClusterName).Should(Equal(expected.Spec.ClusterName))
		Expect(actual.Spec.ClusterId).Should(Equal(expected.Spec.ClusterId))

		Expect(actual.Spec.OBServerTemplate.Image).Should(Equal(expected.Spec.OBServerTemplate.Image))
		Expect(actual.Spec.OBServerTemplate.Resource.Cpu.Value()).Should(Equal(expected.Spec.OBServerTemplate.Resource.Cpu.Value()))
		Expect(actual.Spec.OBServerTemplate.Resource.Memory.Value()).Should(Equal(expected.Spec.OBServerTemplate.Resource.Memory.Value()))
		Expect(actual.Spec.OBServerTemplate.Storage.DataStorage.StorageClass).Should(Equal(expected.Spec.OBServerTemplate.Storage.DataStorage.StorageClass))
		Expect(actual.Spec.OBServerTemplate.Storage.DataStorage.Size.Value()).Should(Equal(expected.Spec.OBServerTemplate.Storage.DataStorage.Size.Value()))
		Expect(actual.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass).Should(Equal(expected.Spec.OBServerTemplate.Storage.RedoLogStorage.StorageClass))
		Expect(actual.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Value()).Should(Equal(expected.Spec.OBServerTemplate.Storage.RedoLogStorage.Size.Value()))
		Expect(actual.Spec.OBServerTemplate.Storage.LogStorage.StorageClass).Should(Equal(expected.Spec.OBServerTemplate.Storage.LogStorage.StorageClass))
		Expect(actual.Spec.OBServerTemplate.Storage.LogStorage.Size.Value()).Should(Equal(expected.Spec.OBServerTemplate.Storage.LogStorage.Size.Value()))

		Expect(actual.Spec.MonitorTemplate.Image).Should(Equal(expected.Spec.MonitorTemplate.Image))
		Expect(actual.Spec.MonitorTemplate.Resource.Cpu.Value()).Should(Equal(expected.Spec.MonitorTemplate.Resource.Cpu.Value()))
		Expect(actual.Spec.MonitorTemplate.Resource.Memory.Value()).Should(Equal(expected.Spec.MonitorTemplate.Resource.Memory.Value()))

		Expect(actual.Spec.BackupVolume.Volume.Name).Should(Equal(expected.Spec.BackupVolume.Volume.Name))
		Expect(actual.Spec.BackupVolume.Volume.VolumeSource.NFS.Server).Should(Equal(expected.Spec.BackupVolume.Volume.VolumeSource.NFS.Server))
		Expect(actual.Spec.BackupVolume.Volume.VolumeSource.NFS.Path).Should(Equal(expected.Spec.BackupVolume.Volume.VolumeSource.NFS.Path))

		Expect(actual.Spec.Parameters).Should(BeEquivalentTo(expected.Spec.Parameters))

		Expect(len(actual.Spec.Topology)).Should(Equal(len(expected.Spec.Topology)))
		for i, zone := range actual.Spec.Topology {
			Expect(zone.Zone).Should(Equal(expected.Spec.Topology[i].Zone))
			Expect(zone.Replica).Should(Equal(expected.Spec.Topology[i].Replica))
			Expect(zone.NodeSelector).Should(BeEquivalentTo(expected.Spec.Topology[i].NodeSelector))
			Expect(zone.Affinity).Should(BeEquivalentTo(expected.Spec.Topology[i].Affinity))
			Expect(zone.Tolerations).Should(BeEquivalentTo(expected.Spec.Topology[i].Tolerations))
		}

		// actual == expected-<uuid>
		Expect(actual.Spec.UserSecrets.Root).Should(HavePrefix(expected.Spec.UserSecrets.Root))
		Expect(actual.Spec.UserSecrets.ProxyRO).Should(HavePrefix(expected.Spec.UserSecrets.ProxyRO))
		Expect(actual.Spec.UserSecrets.Monitor).Should(HavePrefix(expected.Spec.UserSecrets.Monitor))
		Expect(actual.Spec.UserSecrets.Operator).Should(HavePrefix(expected.Spec.UserSecrets.Operator))
	})
})
