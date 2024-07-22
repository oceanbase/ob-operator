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
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/clients/schema"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func buildOBTenantApiType(nn types.NamespacedName, p *param.CreateOBTenantParam) (*v1alpha1.OBTenant, error) {
	t := &v1alpha1.OBTenant{
		ObjectMeta: v1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		TypeMeta: v1.TypeMeta{
			Kind:       schema.OBTenantKind,
			APIVersion: schema.OBTenantGroup + "/" + schema.OBTenantVersion,
		},
		Spec: v1alpha1.OBTenantSpec{
			ClusterName:      p.ClusterName,
			TenantName:       p.TenantName,
			UnitNumber:       p.UnitNumber,
			Charset:          p.Charset,
			ConnectWhiteList: p.ConnectWhiteList,
			TenantRole:       apitypes.TenantRole(p.TenantRole),

			// guard non-nil
			Pools: []v1alpha1.ResourcePoolSpec{},
		},
	}

	if len(p.Pools) == 0 {
		return nil, oberr.NewBadRequest("pools is empty")
	}
	if p.UnitConfig == nil {
		return nil, oberr.NewBadRequest("unit config is nil")
	}

	cpuCount, err := resource.ParseQuantity(p.UnitConfig.CPUCount)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid cpu count: " + err.Error())
	}
	memorySize, err := resource.ParseQuantity(p.UnitConfig.MemorySize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid memory size: " + err.Error())
	}
	logDiskSize, err := resource.ParseQuantity(p.UnitConfig.LogDiskSize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid log disk size: " + err.Error())
	}

	t.Spec.Pools = make([]v1alpha1.ResourcePoolSpec, 0, len(p.Pools))
	for i := range p.Pools {
		apiPool := v1alpha1.ResourcePoolSpec{
			Zone:       p.Pools[i].Zone,
			Priority:   p.Pools[i].Priority,
			Type:       &v1alpha1.LocalityType{},
			UnitConfig: &v1alpha1.UnitConfig{},
		}
		apiPool.Type = &v1alpha1.LocalityType{
			Name:     p.Pools[i].Type,
			Replica:  1,
			IsActive: true,
		}
		apiPool.UnitConfig = &v1alpha1.UnitConfig{
			MaxCPU:      cpuCount,
			MemorySize:  memorySize,
			MinCPU:      cpuCount,
			LogDiskSize: logDiskSize,
			MaxIops:     p.UnitConfig.MaxIops,
			MinIops:     p.UnitConfig.MinIops,
			IopsWeight:  p.UnitConfig.IopsWeight,
		}
		t.Spec.Pools = append(t.Spec.Pools, apiPool)
	}

	if p.Source != nil {
		t.Spec.Source = &v1alpha1.TenantSourceSpec{
			Tenant: p.Source.Tenant,
		}
		if p.Source.Restore != nil {
			t.Spec.Source.Restore = &v1alpha1.RestoreSourceSpec{
				ArchiveSource: &apitypes.BackupDestination{},
				BakDataSource: &apitypes.BackupDestination{},
				// BakEncryptionSecret: p.Source.Restore.BakEncryptionSecret,
				Until: v1alpha1.RestoreUntilConfig{},
			}

			t.Spec.Source.Restore.ArchiveSource.Type = apitypes.BackupDestType(p.Source.Restore.Type)
			t.Spec.Source.Restore.ArchiveSource.Path = p.Source.Restore.ArchiveSource
			t.Spec.Source.Restore.BakDataSource.Type = apitypes.BackupDestType(p.Source.Restore.Type)
			t.Spec.Source.Restore.BakDataSource.Path = p.Source.Restore.BakDataSource

			if p.Source.Restore.Until != nil && !p.Source.Restore.Until.Unlimited {
				t.Spec.Source.Restore.Until.Timestamp = p.Source.Restore.Until.Timestamp
			} else {
				t.Spec.Source.Restore.Until.Unlimited = true
			}
		}
	}
	return t, nil
}

func buildDetailFromApiType(t *v1alpha1.OBTenant) *response.OBTenantDetail {
	rt := &response.OBTenantDetail{
		OBTenantOverview: *buildOverviewFromApiType(t),
	}
	rt.RootCredential = t.Status.Credentials.Root
	rt.StandbyROCredential = t.Status.Credentials.StandbyRO

	if t.Status.Source != nil && t.Status.Source.Tenant != nil {
		rt.PrimaryTenant = *t.Status.Source.Tenant
	}

	if t.Spec.Source != nil && t.Spec.Source.Restore != nil {
		rt.RestoreSource = &response.RestoreSource{
			Type:                string(t.Spec.Source.Restore.ArchiveSource.Type),
			ArchiveSource:       t.Spec.Source.Restore.ArchiveSource.Path,
			BakDataSource:       t.Spec.Source.Restore.BakDataSource.Path,
			OssAccessSecret:     t.Spec.Source.Restore.ArchiveSource.OSSAccessSecret,
			BakEncryptionSecret: t.Spec.Source.Restore.BakEncryptionSecret,
		}
		if !t.Spec.Source.Restore.Until.Unlimited {
			rt.RestoreSource.Until = *t.Spec.Source.Restore.Until.Timestamp
		}
	}

	return rt
}

func buildOverviewFromApiType(t *v1alpha1.OBTenant) *response.OBTenantOverview {
	rt := &response.OBTenantOverview{}
	rt.UID = string(t.UID)
	rt.Name = t.Name
	rt.Namespace = t.Namespace
	rt.CreateTime = t.CreationTimestamp.Format("2006-01-02 15:04:05")
	rt.TenantName = t.Spec.TenantName
	rt.ClusterName = t.Spec.ClusterName
	rt.TenantRole = string(t.Status.TenantRole)
	rt.UnitNumber = t.Spec.UnitNumber
	rt.Status = t.Status.Status
	rt.Charset = t.Spec.Charset
	rt.Locality = t.Status.TenantRecordInfo.Locality
	rt.PrimaryZone = t.Status.TenantRecordInfo.PrimaryZone

	for i := range t.Status.Pools {
		pool := t.Status.Pools[i]
		replica := response.OBTenantReplica{
			Zone:     pool.ZoneList,
			Priority: pool.Priority,
			Type:     pool.Type.Name,
		}
		if pool.UnitConfig != nil {
			replica.MaxCPU = pool.UnitConfig.MaxCPU.String()
			replica.MinCPU = pool.UnitConfig.MinCPU.String()
			replica.MaxIops = pool.UnitConfig.MaxIops
			replica.MinIops = pool.UnitConfig.MinIops
			replica.IopsWeight = pool.UnitConfig.IopsWeight
			replica.MemorySize = fmt.Sprintf("%d GB", pool.UnitConfig.MemorySize.Value()>>30)
			replica.LogDiskSize = fmt.Sprintf("%d GB", pool.UnitConfig.LogDiskSize.Value()>>30)
		}
		rt.Topology = append(rt.Topology, replica)
	}
	return rt
}

func updateOBTenant(ctx context.Context, nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}

	tenant.Spec = t.Spec
	tenant, err = clients.UpdateOBTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	return buildDetailFromApiType(tenant), nil
}

func createPasswordSecret(ctx context.Context, nn types.NamespacedName, password string) error {
	k8sclient := client.GetClient()
	_, err := k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		},
		StringData: map[string]string{
			"password": password,
		},
	}, v1.CreateOptions{})
	return err
}

func CreateOBTenant(ctx context.Context, nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenantDetail, error) {
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}
	if p.RootPassword != "" {
		t.Spec.Credentials.Root = p.Name + "-root-" + rand.String(6)
	}

	k8sclient := client.GetClient()

	if p.Source != nil && p.Source.Tenant != nil {
		// Check primary tenant
		ns := nn.Namespace
		tenantCR := *p.Source.Tenant
		if strings.Contains(*p.Source.Tenant, "/") {
			splits := strings.Split(*p.Source.Tenant, "/")
			if len(splits) != 2 {
				return nil, oberr.NewBadRequest("invalid tenant name")
			}
			ns, tenantCR = splits[0], splits[1]
		}
		existing, err := clients.GetOBTenant(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      tenantCR,
		})
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				return nil, oberr.NewBadRequest("primary tenant not found")
			}
			return nil, oberr.NewInternal(err.Error())
		}
		if existing.Status.TenantRole != apiconst.TenantRolePrimary {
			return nil, oberr.NewBadRequest("the target tenant is not primary tenant")
		}

		// Match root password
		rootSecret, err := k8sclient.ClientSet.CoreV1().Secrets(existing.Namespace).Get(ctx, existing.Status.Credentials.Root, v1.GetOptions{})
		if err != nil {
			return nil, oberr.NewInternal(err.Error())
		}
		if pwd, ok := rootSecret.Data["password"]; ok {
			if p.RootPassword != string(pwd) {
				return nil, oberr.NewBadRequest("root password not match")
			}
			if t.Spec.Credentials.Root != "" {
				err = createPasswordSecret(ctx, types.NamespacedName{
					Namespace: nn.Namespace,
					Name:      t.Spec.Credentials.Root,
				}, p.RootPassword)
				if err != nil {
					return nil, oberr.NewInternal(err.Error())
				}
			}
		}

		// Fetch standbyro password
		standbyroSecret, err := k8sclient.ClientSet.CoreV1().Secrets(existing.Namespace).Get(ctx, existing.Status.Credentials.StandbyRO, v1.GetOptions{})
		if err != nil {
			return nil, oberr.NewInternal(err.Error())
		}
		if pwd, ok := standbyroSecret.Data["password"]; ok {
			t.Spec.Credentials.StandbyRO = p.Name + "-standbyro-" + rand.String(6)
			err = createPasswordSecret(ctx, types.NamespacedName{
				Namespace: nn.Namespace,
				Name:      t.Spec.Credentials.StandbyRO,
			}, string(pwd))
			if err != nil {
				return nil, oberr.NewInternal(err.Error())
			}
		}
	} else {
		if t.Spec.Credentials.Root != "" {
			err = createPasswordSecret(ctx, types.NamespacedName{
				Namespace: nn.Namespace,
				Name:      t.Spec.Credentials.Root,
			}, p.RootPassword)
			if err != nil {
				return nil, oberr.NewInternal(err.Error())
			}
		}
		t.Spec.Credentials.StandbyRO = p.Name + "-standbyro-" + rand.String(6)
		err = createPasswordSecret(ctx, types.NamespacedName{
			Namespace: nn.Namespace,
			Name:      t.Spec.Credentials.StandbyRO,
		}, rand.String(32))
		if err != nil {
			return nil, oberr.NewInternal(err.Error())
		}
	}

	if p.Source != nil && p.Source.Restore != nil {
		if p.Source.Restore.BakEncryptionPassword != "" {
			secretName := p.Name + "-bak-encryption-" + rand.String(6)
			t.Spec.Source.Restore.BakEncryptionSecret = secretName
			_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      secretName,
					Namespace: nn.Namespace,
				},
				StringData: map[string]string{
					"password": p.Source.Restore.BakEncryptionPassword,
				},
			}, v1.CreateOptions{})
			if err != nil {
				return nil, oberr.NewInternal(err.Error())
			}
		}

		if p.Source.Restore.OSSAccessID != "" && p.Source.Restore.OSSAccessKey != "" {
			ossSecretName := p.Name + "-oss-access-" + rand.String(6)
			t.Spec.Source.Restore.ArchiveSource.OSSAccessSecret = ossSecretName
			t.Spec.Source.Restore.BakDataSource.OSSAccessSecret = ossSecretName
			_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      ossSecretName,
					Namespace: nn.Namespace,
				},
				StringData: map[string]string{
					"accessId":  p.Source.Restore.OSSAccessID,
					"accessKey": p.Source.Restore.OSSAccessKey,
				},
			}, v1.CreateOptions{})
			if err != nil {
				return nil, oberr.NewInternal(err.Error())
			}
		}
	}

	tenant, err := clients.CreateOBTenant(ctx, t)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ListAllOBTenants(ctx context.Context, ns string, listOptions v1.ListOptions) ([]*response.OBTenantOverview, error) {
	tenantList, err := clients.ListAllOBTenants(ctx, ns, listOptions)
	if err != nil {
		return nil, err
	}
	sort.Slice(tenantList.Items, func(i, j int) bool {
		return tenantList.Items[i].Name < tenantList.Items[j].Name
	})
	tenants := make([]*response.OBTenantOverview, 0, len(tenantList.Items))
	for i := range tenantList.Items {
		tenants = append(tenants, buildOverviewFromApiType(&tenantList.Items[i]))
	}
	return tenants, nil
}

func GetOBTenant(ctx context.Context, nn types.NamespacedName) (*response.OBTenantDetail, error) {
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func DeleteOBTenant(ctx context.Context, nn types.NamespacedName) error {
	return clients.DeleteOBTenant(ctx, nn)
}

func ModifyOBTenantRootPassword(ctx context.Context, nn types.NamespacedName, rootPassword string) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	// create new secret
	k8sclient := client.GetClient()
	newRootSecretName := nn.Name + "-root-" + rand.String(6)
	_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      newRootSecretName,
			Namespace: nn.Namespace,
		},
		StringData: map[string]string{
			"password": rootPassword,
		},
	}, v1.CreateOptions{})
	if err != nil {
		return nil, oberr.NewInternal(err.Error())
	}

	changePwdOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      nn.Name + "-change-root-pwd-" + rand.String(6),
			Namespace: nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpChangePwd,
			ChangePwd: &v1alpha1.OBTenantOpChangePwdSpec{
				Tenant:    nn.Name,
				SecretRef: newRootSecretName,
			},
		},
	}
	_, err = clients.CreateOBTenantOperation(ctx, &changePwdOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ReplayStandbyLog(ctx context.Context, nn types.NamespacedName, param *param.ReplayStandbyLog) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if tenant.Status.TenantRole != apiconst.TenantRoleStandby {
		return nil, errors.New("The tenant is not standby tenant")
	}
	replayLogOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      nn.Name + "-replay-log-" + rand.String(6),
			Namespace: nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpReplayLog,
			ReplayUntil: &v1alpha1.RestoreUntilConfig{
				Timestamp: param.Timestamp,
				Unlimited: param.Unlimited,
			},
			TargetTenant: &nn.Name,
		},
	}
	_, err = clients.CreateOBTenantOperation(ctx, &replayLogOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func UpgradeTenantVersion(ctx context.Context, nn types.NamespacedName) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if tenant.Status.TenantRole != apiconst.TenantRolePrimary {
		return nil, errors.New("The tenant is not primary tenant")
	}
	upgradeOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      nn.Name + "-upgrade-" + rand.String(6),
			Namespace: nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type:         apiconst.TenantOpUpgrade,
			TargetTenant: &nn.Name,
		},
	}
	_, err = clients.CreateOBTenantOperation(ctx, &upgradeOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ChangeTenantRole(ctx context.Context, nn types.NamespacedName, p *param.ChangeTenantRole) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if tenant.Status.TenantRole == apiconst.TenantRolePrimary && p.Failover {
		return nil, oberr.NewBadRequest("The tenant is already PRIMARY")
	}
	if p.Switchover && (tenant.Status.Source == nil || tenant.Status.Source.Tenant == nil) {
		return nil, oberr.NewBadRequest("The tenant has no primary tenant")
	}
	changeRoleOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			Name:      nn.Name + "-change-role-" + rand.String(6),
			Namespace: nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{},
	}
	if p.Switchover {
		changeRoleOp.Spec.Type = apiconst.TenantOpSwitchover
		changeRoleOp.Spec.Switchover = &v1alpha1.OBTenantOpSwitchoverSpec{
			PrimaryTenant: *tenant.Status.Source.Tenant,
			StandbyTenant: nn.Name,
		}
	} else if p.Failover {
		changeRoleOp.Spec.Type = apiconst.TenantOpFailover
		changeRoleOp.Spec.Failover = &v1alpha1.OBTenantOpFailoverSpec{
			StandbyTenant: nn.Name,
		}
	}
	_, err = clients.CreateOBTenantOperation(ctx, &changeRoleOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func PatchTenant(ctx context.Context, nn types.NamespacedName, p *param.PatchTenant) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if p.UnitNumber != nil {
		tenant.Spec.UnitNumber = *p.UnitNumber
	}
	if p.UnitConfig != nil {
		cpuCount, err := resource.ParseQuantity(p.UnitConfig.UnitConfig.CPUCount)
		if err != nil {
			return nil, oberr.NewBadRequest("invalid cpu count: " + err.Error())
		}
		memorySize, err := resource.ParseQuantity(p.UnitConfig.UnitConfig.MemorySize)
		if err != nil {
			return nil, oberr.NewBadRequest("invalid memory size: " + err.Error())
		}
		logDiskSize, err := resource.ParseQuantity(p.UnitConfig.UnitConfig.LogDiskSize)
		if err != nil {
			return nil, oberr.NewBadRequest("invalid log disk size: " + err.Error())
		}
		for _, pool := range p.UnitConfig.Pools {
			for i := range tenant.Spec.Pools {
				if tenant.Spec.Pools[i].Zone == pool.Zone {
					tenant.Spec.Pools[i].Priority = pool.Priority
					tenant.Spec.Pools[i].Type.Name = pool.Type
					tenant.Spec.Pools[i].UnitConfig = &v1alpha1.UnitConfig{
						MaxCPU:      cpuCount,
						MemorySize:  memorySize,
						MinCPU:      cpuCount,
						IopsWeight:  p.UnitConfig.UnitConfig.IopsWeight,
						MaxIops:     p.UnitConfig.UnitConfig.MaxIops,
						MinIops:     p.UnitConfig.UnitConfig.MinIops,
						LogDiskSize: logDiskSize,
					}
					break
				}
			}
		}
	}
	tenant, err = clients.UpdateOBTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

// GetOBTenantStatistics returns the statistics of all tenants
// Including the number of tenants in four status: running, deleting, operating, failed
func GetOBTenantStatistics(ctx context.Context) ([]response.OBTenantStatistic, error) {
	stats := []response.OBTenantStatistic{}
	tenantList, err := clients.ListAllOBTenants(ctx, "", v1.ListOptions{})
	if err != nil {
		return nil, oberr.Wrap(err, oberr.ErrInternal, "failed to list tenants")
	}
	var runningCount, deletingCount, operatingCount, failedCount int
	for _, tenant := range tenantList.Items {
		switch tenant.Status.Status {
		case tenantstatus.Running:
			runningCount++
		case tenantstatus.DeletingTenant:
			deletingCount++
		case tenantstatus.Failed, tenantstatus.RestoreFailed:
			failedCount++
		default:
			operatingCount++
		}
	}
	stats = append(stats, response.OBTenantStatistic{
		Status: tenantstatus.Running,
		Count:  runningCount,
	}, response.OBTenantStatistic{
		Status: tenantstatus.DeletingTenant,
		Count:  deletingCount,
	}, response.OBTenantStatistic{
		Status: "operating",
		Count:  operatingCount,
	}, response.OBTenantStatistic{
		Status: tenantstatus.Failed,
		Count:  failedCount,
	})
	return stats, nil
}
