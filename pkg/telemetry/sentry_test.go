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

package telemetry

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

var _ = Describe("Telemetry sentry", Label("sentry"), func() {
	It("cluster sentry", func() {
		cluster := &v1alpha1.OBCluster{}
		objectSentry(cluster)

		beforeServer := "1.2.3.4"
		cluster.Spec.BackupVolume = &v1alpha1.BackupVolumeSpec{
			Volume: &corev1.Volume{
				Name: "backup",
				VolumeSource: corev1.VolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Server: beforeServer,
						Path:   "/backup/opt",
					},
				},
			},
		}
		objectSentry(cluster)
		Expect(cluster.Spec.BackupVolume.Volume.NFS.Server).ShouldNot(Equal(beforeServer))
	})

	It("server sentry", func() {
		podIP := "127.0.1.2"
		nodeIP := "128.0.1.2"
		server := &v1alpha1.OBServer{
			Status: v1alpha1.OBServerStatus{
				PodIp:  podIP,
				NodeIp: nodeIP,
			},
		}
		objectSentry(server)
		Expect(server.Status.PodIp).ShouldNot(Equal(podIP))
		Expect(server.Status.NodeIp).ShouldNot(Equal(nodeIP))
	})

	It("tenant sentry", func() {
		beforeIp := "1.2.3.4"
		tenant := &v1alpha1.OBTenant{
			Status: v1alpha1.OBTenantStatus{
				Pools: []v1alpha1.ResourcePoolStatus{{
					Units: []v1alpha1.UnitStatus{{
						UnitId:     0,
						ServerIP:   beforeIp,
						ServerPort: 0,
						Status:     "",
						Migrate: v1alpha1.MigrateServerStatus{
							ServerIP: beforeIp,
						},
					}},
				}},
			},
		}
		objectSentry(tenant)
		Expect(tenant.Status.Pools[0].Units[0].ServerIP).ShouldNot(Equal(beforeIp))
		Expect(tenant.Status.Pools[0].Units[0].Migrate.ServerIP).ShouldNot(Equal(beforeIp))
	})

	It("zone sentry", func() {
		beforeIp := "1.2.3.4"
		zone := &v1alpha1.OBZone{
			Status: v1alpha1.OBZoneStatus{
				OBServerStatus: []v1alpha1.OBServerReplicaStatus{{
					Server: beforeIp,
				}},
			},
			Spec: v1alpha1.OBZoneSpec{
				BackupVolume: &v1alpha1.BackupVolumeSpec{
					Volume: &corev1.Volume{
						Name: "backup",
						VolumeSource: corev1.VolumeSource{
							NFS: &corev1.NFSVolumeSource{
								Server: beforeIp,
								Path:   "/backup/opt",
							},
						},
					},
				},
			},
		}
		objectSentry(zone)
		Expect(zone.Status.OBServerStatus[0].Server).ShouldNot(Equal(beforeIp))
		Expect(zone.Spec.BackupVolume.Volume.NFS.Server).ShouldNot(Equal(beforeIp))
	})
})
