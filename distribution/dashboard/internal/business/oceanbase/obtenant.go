package oceanbase

import (
	"context"
	"errors"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	oberr "github.com/oceanbase/oceanbase-dashboard/pkg/errors"
	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase/schema"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/retry"
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
			Pools:  []v1alpha1.ResourcePoolSpec{},
			Source: &v1alpha1.TenantSourceSpec{},
		},
	}
	if p.RootPassword != "" {
		t.Spec.Credentials.Root = p.Name + "-root-" + rand.String(6)
	}
	t.Spec.Credentials.StandbyRO = p.Name + "-standbyro-" + rand.String(6)

	if len(p.Pools) == 0 {
		return nil, errors.New("pools is empty")
	}
	// if len(p.Pools) > 0 {
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
		if p.UnitConfig != nil {
			apiPool.UnitConfig = &v1alpha1.UnitConfig{
				MaxCPU:      resource.MustParse(p.UnitConfig.CPUCount),
				MemorySize:  resource.MustParse(p.UnitConfig.MemorySize),
				MinCPU:      resource.MustParse(p.UnitConfig.CPUCount),
				LogDiskSize: resource.MustParse(p.UnitConfig.LogDiskSize),
				MaxIops:     p.UnitConfig.MaxIops,
				MinIops:     p.UnitConfig.MinIops,
				IopsWeight:  p.UnitConfig.IopsWeight,
			}
		}
		t.Spec.Pools = append(t.Spec.Pools, apiPool)
	}
	if p.Source != nil {
		t.Spec.Source = &v1alpha1.TenantSourceSpec{
			Tenant:  p.Source.Tenant,
			Restore: &v1alpha1.RestoreSourceSpec{},
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

			if p.Source.Restore.BakEncryptionPassword != "" {
				t.Spec.Credentials.Root = p.Name + "-bak-encryption-" + rand.String(6)
			}

			if p.Source.Restore.OSSAccessID != "" && p.Source.Restore.OSSAccessKey != "" {
				ossName := p.Name + "-oss-access-" + rand.String(6)
				t.Spec.Source.Restore.ArchiveSource.OSSAccessSecret = ossName
				t.Spec.Source.Restore.BakDataSource.OSSAccessSecret = ossName
			}

			if p.Source.Restore.Until != nil {
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
		OBTenantBrief: *buildBriefFromApiType(t),
	}
	rt.RootCredential = t.Status.Credentials.Root
	rt.StandbyROCredentail = t.Status.Credentials.StandbyRO

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

func buildBriefFromApiType(t *v1alpha1.OBTenant) *response.OBTenantBrief {
	rt := &response.OBTenantBrief{}
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

	for i := range t.Spec.Pools {
		pool := t.Spec.Pools[i]
		replica := response.OBTenantReplica{
			Zone:     pool.Zone,
			Priority: pool.Priority,
			Type:     pool.Type.Name,
		}
		if pool.UnitConfig != nil {
			replica.MaxCPU = pool.UnitConfig.MaxCPU.String()
			replica.MemorySize = pool.UnitConfig.MemorySize.String()
			replica.MinCPU = pool.UnitConfig.MinCPU.String()
			replica.MaxIops = pool.UnitConfig.MaxIops
			replica.MinIops = pool.UnitConfig.MinIops
			replica.IopsWeight = pool.UnitConfig.IopsWeight
			replica.LogDiskSize = pool.UnitConfig.LogDiskSize.String()
		}
		rt.Topology = append(rt.Topology, replica)
	}
	return rt
}

func CreateOBTenant(ctx context.Context, nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenantDetail, error) {
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}
	if t.Spec.Credentials.Root != "" {
		k8sclient := client.GetClient()
		_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(context.TODO(), &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      t.Spec.Credentials.Root,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"password": p.RootPassword,
			},
		}, v1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	if t.Spec.Credentials.StandbyRO != "" {
		k8sclient := client.GetClient()
		_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(context.TODO(), &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      t.Spec.Credentials.StandbyRO,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"password": rand.String(20),
			},
		}, v1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	if t.Spec.Source != nil && t.Spec.Source.Restore != nil && t.Spec.Source.Restore.BakEncryptionSecret != "" &&
		p.Source != nil && p.Source.Restore != nil && p.Source.Restore.BakEncryptionPassword != "" {
		k8sclient := client.GetClient()
		_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(context.TODO(), &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      t.Spec.Credentials.Root,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"password": p.Source.Restore.BakEncryptionPassword,
			},
		}, v1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	tenant, err := oceanbase.CreateOBTenant(ctx, t)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func updateOBTenant(ctx context.Context, nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		tenant, err := oceanbase.GetOBTenant(ctx, nn)
		if err != nil {
			return err
		}
		tenant.Spec = t.Spec
		tenant, err = oceanbase.UpdateOBTenant(ctx, tenant)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ListAllOBTenants(ctx context.Context, listOptions metav1.ListOptions) ([]*response.OBTenantBrief, error) {
	tenantList, err := oceanbase.ListAllOBTenants(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	tenants := make([]*response.OBTenantBrief, 0, len(tenantList.Items))
	for i := range tenantList.Items {
		tenants = append(tenants, buildBriefFromApiType(&tenantList.Items[i]))
	}
	return tenants, nil
}

func GetOBTenant(ctx context.Context, nn types.NamespacedName) (*response.OBTenantDetail, error) {
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func DeleteOBTenant(ctx context.Context, nn types.NamespacedName) error {
	return oceanbase.DeleteOBTenant(ctx, nn)
}

func ModifyOBTenantUnitNumber(ctx context.Context, nn types.NamespacedName, unitNumber int) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}

	tenant.Spec.UnitNumber = unitNumber
	tenant, err = oceanbase.UpdateOBTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ModifyOBTenantUnitConfig(ctx context.Context, nn types.NamespacedName, zone string, unitConfig *param.UnitConfig) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	for i := range tenant.Spec.Pools {
		if tenant.Spec.Pools[i].Zone == zone {
			tenant.Spec.Pools[i].UnitConfig = &v1alpha1.UnitConfig{
				MaxCPU:      resource.MustParse(unitConfig.CPUCount),
				MemorySize:  resource.MustParse(unitConfig.MemorySize),
				MinCPU:      resource.MustParse(unitConfig.CPUCount),
				LogDiskSize: resource.MustParse(unitConfig.LogDiskSize),
				MaxIops:     unitConfig.MaxIops,
				MinIops:     unitConfig.MinIops,
				IopsWeight:  unitConfig.IopsWeight,
			}
			break
		}
	}
	tenant, err = oceanbase.UpdateOBTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ModifyOBTenantRootPassword(ctx context.Context, nn types.NamespacedName, rootPassword string) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	// create new secret
	k8sclient := client.GetClient()
	newRootSecretName := nn.Name + "-root-" + rand.String(6)
	_, err = k8sclient.ClientSet.CoreV1().Secrets(nn.Namespace).Create(context.TODO(), &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      newRootSecretName,
			Namespace: nn.Namespace,
		},
		StringData: map[string]string{
			"password": rootPassword,
		},
	}, v1.CreateOptions{})

	changePwdOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: nn.Name + "-change-root-pwd-",
			Namespace:    nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpChangePwd,
			ChangePwd: &v1alpha1.OBTenantOpChangePwdSpec{
				Tenant:    nn.Name,
				SecretRef: newRootSecretName,
			},
		},
	}
	_, err = oceanbase.CreateOBTenantOperation(ctx, &changePwdOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ReplayStandbyLog(ctx context.Context, nn types.NamespacedName, timestamp string) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if tenant.Status.TenantRole != apiconst.TenantRoleStandby {
		return nil, errors.New("The tenant is not standby tenant")
	}
	replayLogOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: nn.Name + "-replay-log-",
			Namespace:    nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type: apiconst.TenantOpReplayLog,
			ReplayUntil: &v1alpha1.RestoreUntilConfig{
				Timestamp: &timestamp,
			},
			TargetTenant: &nn.Name,
		},
	}
	_, err = oceanbase.CreateOBTenantOperation(ctx, &replayLogOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func UpgradeTenantVersion(ctx context.Context, nn types.NamespacedName) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if tenant.Status.TenantRole != apiconst.TenantRolePrimary {
		return nil, errors.New("The tenant is not primary tenant")
	}
	upgradeOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: nn.Name + "-upgrade-",
			Namespace:    nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{
			Type:         apiconst.TenantOpUpgrade,
			TargetTenant: &nn.Name,
		},
	}
	_, err = oceanbase.CreateOBTenantOperation(ctx, &upgradeOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func ChangeTenantRole(ctx context.Context, nn types.NamespacedName, p *param.ChangeTenantRole) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if tenant.Status.TenantRole == apitypes.TenantRole(p.TenantRole) {
		return nil, oberr.NewBadRequest("The tenant is already " + string(p.TenantRole))
	}
	if p.Switchover && (tenant.Status.Source == nil || tenant.Status.Source.Tenant == nil) {
		return nil, oberr.NewBadRequest("The tenant has no primary tenant")
	}
	changeRoleOp := v1alpha1.OBTenantOperation{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: nn.Name + "-change-role-",
			Namespace:    nn.Namespace,
		},
		Spec: v1alpha1.OBTenantOperationSpec{},
	}
	if p.Switchover {
		changeRoleOp.Spec.Type = apiconst.TenantOpSwitchover
		changeRoleOp.Spec.Switchover.PrimaryTenant = *tenant.Status.Source.Tenant
		changeRoleOp.Spec.Switchover.StandbyTenant = nn.Name
	} else {
		changeRoleOp.Spec.Type = apiconst.TenantOpFailover
		changeRoleOp.Spec.Failover.StandbyTenant = nn.Name
	}
	_, err = oceanbase.CreateOBTenantOperation(ctx, &changeRoleOp)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func PatchTenant(ctx context.Context, nn types.NamespacedName, p *param.PatchTenant) (*response.OBTenantDetail, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		return nil, err
	}
	if p.UnitNumber != nil {
		tenant.Spec.UnitNumber = *p.UnitNumber
	}
	if p.UnitConfig != nil {
		for _, pool := range p.UnitConfig.Pools {
			for i := range tenant.Spec.Pools {
				if tenant.Spec.Pools[i].Zone == pool.Zone {
					tenant.Spec.Pools[i].Priority = pool.Priority
					tenant.Spec.Pools[i].Type.Name = pool.Type
					tenant.Spec.Pools[i].UnitConfig = &v1alpha1.UnitConfig{
						MaxCPU:      resource.MustParse(p.UnitConfig.UnitConfig.CPUCount),
						MemorySize:  resource.MustParse(p.UnitConfig.UnitConfig.MemorySize),
						MinCPU:      resource.MustParse(p.UnitConfig.UnitConfig.CPUCount),
						IopsWeight:  p.UnitConfig.UnitConfig.IopsWeight,
						MaxIops:     p.UnitConfig.UnitConfig.MaxIops,
						MinIops:     p.UnitConfig.UnitConfig.MinIops,
						LogDiskSize: resource.MustParse(p.UnitConfig.UnitConfig.LogDiskSize),
					}
					break
				}
			}
		}
	}
	tenant, err = oceanbase.UpdateOBTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}
