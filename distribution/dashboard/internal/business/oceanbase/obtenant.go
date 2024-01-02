package oceanbase

import (
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase/schema"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
			ForceDelete:      p.ForceDelete,
			Charset:          p.Charset,
			Collate:          p.Collate,
			ConnectWhiteList: p.ConnectWhiteList,
			TenantRole:       apitypes.TenantRole(p.TenantRole),
			Credentials:      v1alpha1.TenantCredentials(p.Credentials),
			// guard non-nil
			Pools:  []v1alpha1.ResourcePoolSpec{},
			Source: &v1alpha1.TenantSourceSpec{},
		},
	}
	if len(p.Pools) > 0 {
		t.Spec.Pools = make([]v1alpha1.ResourcePoolSpec, 0, len(p.Pools))
		for i := range p.Pools {
			apiPool := v1alpha1.ResourcePoolSpec{
				Zone:       p.Pools[i].Zone,
				Priority:   p.Pools[i].Priority,
				Type:       &v1alpha1.LocalityType{},
				UnitConfig: &v1alpha1.UnitConfig{},
			}
			if p.Pools[i].Type != nil {
				apiPool.Type = &v1alpha1.LocalityType{
					Name:     p.Pools[i].Type.Name,
					Replica:  p.Pools[i].Type.Replica,
					IsActive: p.Pools[i].Type.IsActive,
				}
			}
			if p.Pools[i].UnitConfig != nil {
				apiPool.UnitConfig = &v1alpha1.UnitConfig{
					MaxCPU:      resource.MustParse(p.Pools[i].UnitConfig.MaxCPU),
					MemorySize:  resource.MustParse(p.Pools[i].UnitConfig.MemorySize),
					MinCPU:      resource.MustParse(p.Pools[i].UnitConfig.MinCPU),
					LogDiskSize: resource.MustParse(p.Pools[i].UnitConfig.LogDiskSize),
					MaxIops:     p.Pools[i].UnitConfig.MaxIops,
					MinIops:     p.Pools[i].UnitConfig.MinIops,
					IopsWeight:  p.Pools[i].UnitConfig.IopsWeight,
				}
			}
			t.Spec.Pools = append(t.Spec.Pools, apiPool)
		}
	}
	if p.Source != nil {
		t.Spec.Source = &v1alpha1.TenantSourceSpec{
			Tenant:  p.Source.Tenant,
			Restore: &v1alpha1.RestoreSourceSpec{},
		}
		if p.Source.Restore != nil {
			t.Spec.Source.Restore = &v1alpha1.RestoreSourceSpec{
				ArchiveSource:       &apitypes.BackupDestination{},
				BakDataSource:       &apitypes.BackupDestination{},
				BakEncryptionSecret: p.Source.Restore.BakEncryptionSecret,

				SourceUri:      p.Source.Restore.SourceUri,
				Until:          v1alpha1.RestoreUntilConfig(p.Source.Restore.Until),
				Description:    p.Source.Restore.Description,
				ReplayLogUntil: &v1alpha1.RestoreUntilConfig{},
				Cancel:         p.Source.Restore.Cancel,
			}
			if p.Source.Restore.ArchiveSource != nil {
				t.Spec.Source.Restore.ArchiveSource = &apitypes.BackupDestination{
					Path:            p.Source.Restore.ArchiveSource.Path,
					Type:            apitypes.BackupDestType(p.Source.Restore.ArchiveSource.Type),
					OSSAccessSecret: p.Source.Restore.ArchiveSource.OSSAccessSecret,
				}
			}
			if p.Source.Restore.BakDataSource != nil {
				t.Spec.Source.Restore.BakDataSource = &apitypes.BackupDestination{
					Path:            p.Source.Restore.BakDataSource.Path,
					Type:            apitypes.BackupDestType(p.Source.Restore.BakDataSource.Type),
					OSSAccessSecret: p.Source.Restore.BakDataSource.OSSAccessSecret,
				}
			}
			if p.Source.Restore.ReplayLogUntil != nil {
				t.Spec.Source.Restore.ReplayLogUntil = &v1alpha1.RestoreUntilConfig{
					Timestamp: p.Source.Restore.ReplayLogUntil.Timestamp,
					Scn:       p.Source.Restore.ReplayLogUntil.Scn,
					Unlimited: p.Source.Restore.ReplayLogUntil.Unlimited,
				}
			}
		}
	}
	return t, nil
}

func buildResponseFromApiType(t *v1alpha1.OBTenant) *response.OBTenant {
	rt := &response.OBTenant{}
	rt.Name = t.Name
	rt.Namespace = t.Namespace
	rt.CreatedAt = t.CreationTimestamp.Format("2006-01-02 15:04:05")
	rt.Spec = t.Spec
	rt.Status = t.Status
	return rt
}

func CreateOBTenant(nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenant, error) {
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}
	tenant, err := oceanbase.CreateOBTenant(t)
	if err != nil {
		return nil, err
	}
	return buildResponseFromApiType(tenant), nil
}

func UpdateOBTenant(nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenant, error) {
	var err error
	tenant, err := oceanbase.GetOBTenant(nn)
	if err != nil {
		return nil, err
	}
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		tenant, err := oceanbase.GetOBTenant(nn)
		if err != nil {
			return err
		}
		tenant.Spec = t.Spec
		tenant, err = oceanbase.UpdateOBTenant(tenant)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return buildResponseFromApiType(tenant), nil
}

func ListAllOBTenants() ([]*response.OBTenant, error) {
	tenantList, err := oceanbase.ListAllOBTenants()
	if err != nil {
		return nil, err
	}
	tenants := make([]*response.OBTenant, 0, len(tenantList.Items))
	for i := range tenantList.Items {
		tenants = append(tenants, buildResponseFromApiType(&tenantList.Items[i]))
	}
	return tenants, nil
}

func GetOBTenant(nn types.NamespacedName) (*response.OBTenant, error) {
	tenant, err := oceanbase.GetOBTenant(nn)
	if err != nil {
		return nil, err
	}
	return buildResponseFromApiType(tenant), nil
}

func DeleteOBTenant(nn types.NamespacedName) error {
	return oceanbase.DeleteOBTenant(nn)
}
