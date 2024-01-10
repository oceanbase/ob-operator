package oceanbase

import (
	"errors"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase/schema"
	"k8s.io/apimachinery/pkg/api/resource"
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
				MaxCPU:      resource.MustParse(p.UnitConfig.MaxCPU),
				MemorySize:  resource.MustParse(p.UnitConfig.MemorySize),
				MinCPU:      resource.MustParse(p.UnitConfig.MinCPU),
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

	rt.RootCredential = t.Spec.Credentials.Root
	rt.StandbyROCredentail = t.Spec.Credentials.StandbyRO

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

func CreateOBTenant(nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenantDetail, error) {
	t, err := buildOBTenantApiType(nn, p)
	if err != nil {
		return nil, err
	}
	tenant, err := oceanbase.CreateOBTenant(t)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func UpdateOBTenant(nn types.NamespacedName, p *param.CreateOBTenantParam) (*response.OBTenantDetail, error) {
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
	return buildDetailFromApiType(tenant), nil
}

func ListAllOBTenants(labelSelector string) ([]*response.OBTenantBrief, error) {
	tenantList, err := oceanbase.ListAllOBTenants(labelSelector)
	if err != nil {
		return nil, err
	}
	tenants := make([]*response.OBTenantBrief, 0, len(tenantList.Items))
	for i := range tenantList.Items {
		tenants = append(tenants, buildBriefFromApiType(&tenantList.Items[i]))
	}
	return tenants, nil
}

func GetOBTenant(nn types.NamespacedName) (*response.OBTenantDetail, error) {
	tenant, err := oceanbase.GetOBTenant(nn)
	if err != nil {
		return nil, err
	}
	return buildDetailFromApiType(tenant), nil
}

func DeleteOBTenant(nn types.NamespacedName) error {
	return oceanbase.DeleteOBTenant(nn)
}
