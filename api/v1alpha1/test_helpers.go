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

package v1alpha1

import (
	"fmt"

	apiconsts "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const GB = 1 << 30

const (
	wrongKeySecret      = "wrong-key-secret"
	defaultSecretName   = "test-secret"
	ossAccessSecret     = "oss-access-secret"
	defaultNamespace    = "default"
	defaultStorageClass = "local-path"
	defaultImage        = "oceanbasedev/oceanbase-cn:4.1.0.1-test"
	upgradeImage        = "oceanbasedev/oceanbase-cn:4.2.0.0-test"
)

func newOBCluster(name string, zoneNum int, serverNum int) *OBCluster {
	observerResource := &apitypes.ResourceSpec{
		Cpu:    *resource.NewQuantity(2, resource.DecimalSI),
		Memory: *resource.NewQuantity(10*GB, resource.BinarySI),
	}
	observerStorage := &apitypes.OceanbaseStorageSpec{
		DataStorage: &apitypes.StorageSpec{
			StorageClass: defaultStorageClass,
			Size:         *resource.NewQuantity(50*GB, resource.BinarySI),
		},
		RedoLogStorage: &apitypes.StorageSpec{
			StorageClass: defaultStorageClass,
			Size:         *resource.NewQuantity(50*GB, resource.BinarySI),
		},
		LogStorage: &apitypes.StorageSpec{
			StorageClass: defaultStorageClass,
			Size:         *resource.NewQuantity(10*GB, resource.BinarySI),
		},
	}

	observerTemplate := &apitypes.OBServerTemplate{
		Image:    defaultImage,
		Resource: observerResource,
		Storage:  observerStorage,
	}

	topology := make([]apitypes.OBZoneTopology, zoneNum)
	for i := 0; i < zoneNum; i++ {
		zoneTopology := apitypes.OBZoneTopology{
			Zone:    fmt.Sprintf("zone%d", i),
			Replica: serverNum,
		}
		topology[i] = zoneTopology
	}

	userSecrets := &apitypes.OBUserSecrets{
		Root:     defaultSecretName,
		ProxyRO:  defaultSecretName,
		Monitor:  defaultSecretName,
		Operator: defaultSecretName,
	}

	obcluster := &OBCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   defaultNamespace,
			Annotations: map[string]string{},
		},
		Spec: OBClusterSpec{
			ClusterName:      name,
			ClusterId:        1,
			OBServerTemplate: observerTemplate,
			Topology:         topology,
			UserSecrets:      userSecrets,
		},
	}
	return obcluster
}

func newClusterSecret(name string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
		},
		Data: map[string][]byte{
			"password": []byte(rand.String(16)),
		},
	}
}

func newFakeStorageClass(name string) *storagev1.StorageClass {
	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
		},
		Provisioner: "kubernetes.io/no-provisioner",
	}
}

func newOBTenant(name, clusterName string) *OBTenant {
	return &OBTenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
		},
		Spec: OBTenantSpec{
			ClusterName:      clusterName,
			TenantName:       "t1",
			UnitNumber:       1,
			Charset:          "utf8mb4",
			ConnectWhiteList: "%",
			ForceDelete:      true,
			Credentials: TenantCredentials{
				Root:      defaultSecretName,
				StandbyRO: defaultSecretName,
			},
			Pools: []ResourcePoolSpec{{
				Zone: "zone0",
				Type: &LocalityType{
					Name:     "Full",
					Replica:  1,
					IsActive: true,
				},
				UnitConfig: &UnitConfig{
					MaxCPU:      resource.MustParse("1"),
					MemorySize:  resource.MustParse("5Gi"),
					MinCPU:      resource.MustParse("1"),
					MaxIops:     1024,
					MinIops:     1024,
					IopsWeight:  2,
					LogDiskSize: resource.MustParse("12Gi"),
				},
			}},
		},
	}
}

func newBackupPolicy(policyName, tenantName, clusterName string) *OBTenantBackupPolicy {
	return &OBTenantBackupPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      policyName,
		},
		Spec: OBTenantBackupPolicySpec{
			ObClusterName: clusterName,
			TenantName:    "t1",
			TenantSecret:  defaultSecretName,
			TenantCRName:  tenantName,
			JobKeepWindow: "1d",
			Suspend:       false,
			LogArchive: LogArchiveConfig{
				Destination: apitypes.BackupDestination{
					Path:            "oss://operator-backup-data/archive-t1?host=oss-cn-hangzhou.aliyuncs.com",
					Type:            "OSS",
					OSSAccessSecret: ossAccessSecret,
				},
				SwitchPieceInterval: "1d",
			},
			DataBackup: DataBackupConfig{
				Destination: apitypes.BackupDestination{
					Path:            "oss://operator-backup-data/backup-t1?host=oss-cn-hangzhou.aliyuncs.com",
					Type:            "OSS",
					OSSAccessSecret: ossAccessSecret,
				},
				FullCrontab:        "* * * * *",
				IncrementalCrontab: "* * * * *",
				EncryptionSecret:   defaultSecretName,
			},
			DataClean: CleanPolicy{
				RecoveryWindow: "7d",
			},
		},
	}
}

func newTenantOperation(tenantName string) *OBTenantOperation {
	return &OBTenantOperation{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      rand.String(32),
		},
		Spec: OBTenantOperationSpec{
			Type:            apiconsts.TenantOpChangePwd,
			Switchover:      &OBTenantOpSwitchoverSpec{},
			Failover:        &OBTenantOpFailoverSpec{},
			ChangePwd:       &OBTenantOpChangePwdSpec{},
			ReplayUntil:     &RestoreUntilConfig{},
			TargetTenant:    &tenantName,
			AuxiliaryTenant: &tenantName,
		},
	}
}
