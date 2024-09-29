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
package tenant

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/cli/utils"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/clients/schema"
	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		UnitConfig: &param.UnitConfig{},
		Pools:      make([]param.ResourcePoolSpec, 0),
		Source: &param.TenantSourceSpec{
			Restore: &param.RestoreSourceSpec{
				Until: &param.RestoreUntilConfig{},
			},
		},
		ZonePriority: make(map[string]string),
	}
}

type CreateOptions struct {
	generic.ResourceOptions
	ClusterName      string `json:"obcluster" binding:"required"`
	TenantName       string `json:"tenantName" binding:"required"`
	UnitNumber       int    `json:"unitNum" binding:"required"`
	RootPassword     string `json:"rootPassword" binding:"required"`
	ConnectWhiteList string `json:"connectWhiteList,omitempty"`
	Charset          string `json:"charset,omitempty"`

	UnitConfig *param.UnitConfig        `json:"unitConfig" binding:"required"`
	Pools      []param.ResourcePoolSpec `json:"pools" binding:"required"`

	// Enum: Primary, Standby
	TenantRole string                  `json:"tenantRole,omitempty"`
	Source     *param.TenantSourceSpec `json:"source,omitempty"`

	// Flags for cli
	From         string            `json:"from,omitempty"`
	ZonePriority map[string]string `json:"zones"`
	Restore      bool              `json:"restore"`
	RestoreType  string            `json:"restoreType"`
	Timestamp    string            `json:"timestamp"`
}

func (o *CreateOptions) Parse(_ *cobra.Command, args []string) error {
	pools, err := utils.MapZonesToPools(o.ZonePriority)
	if err != nil {
		return err
	}
	o.Pools = pools
	o.Name = args[0]
	if o.CheckIfFlagChanged("from") {
		o.Source.Tenant = &o.From
		o.TenantRole = "STANDBY"
	} else {
		o.TenantRole = "PRIMARY"
	}
	// create empty standby tenant
	if !o.Restore {
		o.Source.Restore = nil
	}
	return nil
}

func (o *CreateOptions) Complete() error {
	if o.RootPassword == "" {
		o.RootPassword = utils.GenerateRandomPassword(8, 32)
	}
	if o.Timestamp != "" {
		o.Source.Restore.Until.Timestamp = &o.Timestamp
	}
	return nil
}

func (o *CreateOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace is not specified")
	}
	if o.ClusterName == "" {
		return errors.New("cluster name is not specified")
	}
	if o.TenantName == "" {
		return errors.New("tenant name is not specified")
	}
	if !utils.CheckResourceName(o.Name) {
		return fmt.Errorf("invalid resource name in k8s: %s", o.Name)
	}
	if !utils.CheckTenantName(o.TenantName) {
		return fmt.Errorf("invalid tenant name: %s, the first letter must be a letter or an underscore and cannot contain -", o.TenantName)
	}
	if o.Source != nil && o.Source.Tenant != nil && o.TenantRole == "PRIMARY" {
		return fmt.Errorf("invalid tenant role")
	}
	if o.Restore && o.RestoreType != "OSS" && o.RestoreType != "NFS" {
		return errors.New("Restore Type not supported")
	}
	if o.Restore && o.RestoreType == "OSS" && o.Source.Restore.OSSAccessKey == "" {
		return errors.New("oss access key not specified")
	}
	if o.Restore && o.RestoreType == "NFS" && o.Source.Restore.BakEncryptionPassword == "" {
		return errors.New("back encryption password not specified")
	}
	return nil
}

// CreateOBTenant create an obtenant with configs
func CreateOBTenant(ctx context.Context, p *CreateOptions) (*v1alpha1.OBTenant, error) {
	nn := types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}
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
	// Restore
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
	return tenant, nil
}

func buildOBTenantApiType(nn types.NamespacedName, p *CreateOptions) (*v1alpha1.OBTenant, error) {
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
	var maxIops, minIops int
	if p.UnitConfig.MaxIops > math.MaxInt32 {
		maxIops = math.MaxInt32
	} else {
		maxIops = int(p.UnitConfig.MaxIops)
	}
	if p.UnitConfig.MinIops > math.MaxInt32 {
		minIops = math.MaxInt32
	} else {
		minIops = int(p.UnitConfig.MinIops)
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
			MaxIops:     maxIops,
			MinIops:     minIops,
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

			t.Spec.Source.Restore.ArchiveSource.Type = apitypes.BackupDestType(p.RestoreType)
			t.Spec.Source.Restore.ArchiveSource.Path = p.Source.Restore.ArchiveSource
			t.Spec.Source.Restore.BakDataSource.Type = apitypes.BackupDestType(p.RestoreType)
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

// AddFlags for create options
func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	o.AddBaseFlags(cmd)
	o.AddUnitFlags(cmd)
	o.AddPoolFlags(cmd)
	o.AddRestoreFlags(cmd)
}

// AddBaseFlags add base flags
func (o *CreateOptions) AddBaseFlags(cmd *cobra.Command) {
	baseFlags := cmd.Flags()
	baseFlags.StringVarP(&o.TenantName, "tenant-name", "n", "", "Tenant name, if not specified, use name in k8s instead")
	baseFlags.StringVar(&o.ClusterName, "cluster", "", "The cluster name tenant belonged to in k8s")
	baseFlags.StringVar(&o.Namespace, "namespace", "default", "The namespace of the tenant")
	baseFlags.StringVarP(&o.RootPassword, "root-password", "p", "", "The root password of the cluster")
	baseFlags.StringVar(&o.Charset, "charset", "utf8mb4", "The charset using in ob tenant")
	baseFlags.StringVar(&o.ConnectWhiteList, "connect-white-list", "%", "The connect white list using in ob tenant")
	baseFlags.StringVar(&o.From, "from", "", "restore from data source")
}

// AddPoolFlags add pool-related flags
func (o *CreateOptions) AddPoolFlags(cmd *cobra.Command) {
	poolFlags := pflag.NewFlagSet("zone", pflag.ContinueOnError)
	poolFlags.StringToStringVar(&o.ZonePriority, "priority", map[string]string{"z1": "1"}, "The zones of the tenant in the format 'Zone=Priority', multiple values can be provided separated by commas")
	cmd.Flags().AddFlagSet(poolFlags)
}

// AddUnitFlags add unit-resource-related flags
func (o *CreateOptions) AddUnitFlags(cmd *cobra.Command) {
	unitFlags := pflag.NewFlagSet("unit", pflag.ContinueOnError)
	unitFlags.IntVar(&o.UnitNumber, "unit-number", 1, "unit number of the OBTenant")
	unitFlags.Int64Var(&o.UnitConfig.MaxIops, "max-iops", 1024, "The max iops of unit")
	unitFlags.Int64Var(&o.UnitConfig.MinIops, "min-iops", 1024, "The min iops of unit")
	unitFlags.IntVar(&o.UnitConfig.IopsWeight, "iops-weight", 1, "The iops weight of unit")
	unitFlags.StringVar(&o.UnitConfig.CPUCount, "cpu-count", "1", "The cpu count of unit")
	unitFlags.StringVar(&o.UnitConfig.MemorySize, "memory-size", "2Gi", "The memory size of unit")
	unitFlags.StringVar(&o.UnitConfig.LogDiskSize, "log-disk-size", "4Gi", "The log disk size of unit")
	cmd.Flags().AddFlagSet(unitFlags)
}

// AddRestoreFlags add restore flags
func (o *CreateOptions) AddRestoreFlags(cmd *cobra.Command) {
	restoreFlags := pflag.NewFlagSet("restore", pflag.ContinueOnError)
	restoreFlags.BoolVarP(&o.Restore, "restore", "r", false, "Restore from backup files")
	restoreFlags.StringVar(&o.RestoreType, "type", "OSS", "The type of restore source, support OSS or NFS")
	restoreFlags.StringVar(&o.Source.Restore.ArchiveSource, "archive-source", "demo_tenant/log_archive_custom", "The archive source of restore")
	restoreFlags.StringVar(&o.Source.Restore.BakEncryptionPassword, "bak-encryption-password", "", "The backup encryption password of obtenant")
	restoreFlags.StringVar(&o.Source.Restore.BakDataSource, "bak-data-source", "demo_tenant/data_backup_custom_enc", "The bak data source of restore")
	restoreFlags.StringVar(&o.Source.Restore.OSSAccessID, "oss-access-id", "", "The oss access id of restore")
	restoreFlags.StringVar(&o.Source.Restore.OSSAccessKey, "oss-access-key", "", "The oss access key of restore")
	restoreFlags.BoolVar(&o.Source.Restore.Until.Unlimited, "until-unlimited", true, "time limited for restore")
	restoreFlags.StringVar(&o.Timestamp, "until-timestamp", "", "timestamp for obtenant restore")
	cmd.Flags().AddFlagSet(restoreFlags)
}
