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
package backup

import (
	"context"
	"errors"
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/cli/cmd/util"
	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type CreateOptions struct {
	generic.ResourceOption
	BackupPolicyBase
	OSSAccessID           string `json:"ossAccessId,omitempty" example:"encryptedPassword"`
	OSSAccessKey          string `json:"ossAccessKey,omitempty" example:"encryptedPassword"`
	BakEncryptionPassword string `json:"bakEncryptionPassword,omitempty" example:"encryptedPassword"`
}
type BackupPolicyBase struct {
	DestType           string `json:"destType" binding:"required"`
	ArchivePath        string `json:"archivePath" binding:"required"`
	BakDataPath        string `json:"bakDataPath" binding:"required"`
	IncrementalCrontab string `json:"incremental,omitempty"`
	FullCrontab        string `json:"full,omitempty"`
	JobKeepDays        int    `json:"jobKeepDays,omitempty" example:"5"`
	RecoveryDays       int    `json:"recoveryDays,omitempty" example:"3"`
}

// checkCrontabSyntax checks the syntax of the crontab
func checkCrontabSyntax(crontab string) bool {
	if _, err := cron.ParseStandard(crontab); err != nil {
		return false
	}
	return true
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

func buildBackupPolicyApiType(nn types.NamespacedName, obcluster string, p *CreateOptions) (*v1alpha1.OBTenantBackupPolicy, error) {
	policy := &v1alpha1.OBTenantBackupPolicy{}
	policy.Name = nn.Name + "-backup-policy"
	policy.Namespace = nn.Namespace
	policy.Spec = v1alpha1.OBTenantBackupPolicySpec{
		ObClusterName: obcluster,
		TenantCRName:  nn.Name,
		JobKeepWindow: numberToDay(p.JobKeepDays),
		LogArchive: v1alpha1.LogArchiveConfig{
			Destination: apitypes.BackupDestination{
				Path: p.ArchivePath,
				Type: apitypes.BackupDestType(p.DestType),
			},
			SwitchPieceInterval: "1d",
		},
		DataBackup: v1alpha1.DataBackupConfig{
			Destination: apitypes.BackupDestination{
				Path: p.BakDataPath,
				Type: apitypes.BackupDestType(p.DestType),
			},
			FullCrontab:        p.FullCrontab,
			IncrementalCrontab: p.IncrementalCrontab,
		},
		DataClean: v1alpha1.CleanPolicy{
			RecoveryWindow: numberToDay(p.RecoveryDays),
		},
	}
	return policy, nil
}

func CreateTenantBackupPolicy(ctx context.Context, o *CreateOptions) (*v1alpha1.OBTenantBackupPolicy, error) {
	nn := types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
	tenant, err := clients.GetOBTenant(ctx, nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, errors.New("Tenant not found")
		}
		return nil, err
	}
	// check tenant status
	if err := util.CheckTenantStatus(tenant); err != nil {
		return nil, err
	}
	// Check if backup policy already exists
	if backupPolicy, _ := clients.GetTenantBackupPolicy(ctx, nn); backupPolicy != nil {
		return nil, errors.New("Backup policy already exists")
	}
	backupPolicy, err := buildBackupPolicyApiType(nn, tenant.Spec.ClusterName, o)
	if err != nil {
		return nil, err
	}
	if o.DestType == "OSS" && o.OSSAccessID != "" && o.OSSAccessKey != "" {
		ossSecretName := nn.Name + "-backup-oss-secret-" + rand.String(6)
		backupPolicy.Spec.LogArchive.Destination.OSSAccessSecret = ossSecretName
		backupPolicy.Spec.DataBackup.Destination.OSSAccessSecret = ossSecretName
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ossSecretName,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"accessId":  o.OSSAccessID,
				"accessKey": o.OSSAccessKey,
			},
		}
		_, err := client.GetClient().ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	if o.BakEncryptionPassword != "" {
		encryptionSecretName := nn.Name + "-backup-encryption-secret-" + rand.String(6)
		backupPolicy.Spec.DataBackup.EncryptionSecret = encryptionSecretName
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      encryptionSecretName,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"password": o.BakEncryptionPassword,
			},
		}
		_, err := client.GetClient().ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	blockOwnerDeletion := true
	backupPolicy.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion:         tenant.APIVersion,
		Kind:               tenant.Kind,
		Name:               tenant.GetObjectMeta().GetName(),
		UID:                tenant.GetObjectMeta().GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
	}})
	// set labels for backup policy
	backupPolicy.Labels = map[string]string{
		oceanbaseconst.LabelRefOBServer: string(tenant.Spec.ClusterName),
	}
	policy, err := clients.CreateTenantBackupPolicy(ctx, backupPolicy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

func (o *CreateOptions) Complete() error {
	// set default values for archive path and backup data path
	if o.DestType == "NFS" && o.ArchivePath == "" {
		o.ArchivePath = fmt.Sprintf("%s/%s", "archive", o.Name)
	}
	if o.DestType == "NFS" && o.BakDataPath == "" {
		o.BakDataPath = fmt.Sprintf("%s/%s", "backup", o.Name)
	}
	return nil
}

func (o *CreateOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("Namespace is required")
	}
	if o.ArchivePath == "" {
		return errors.New("Archive path is required")
	}
	if o.BakDataPath == "" {
		return errors.New("Backup data path is required")
	}
	if o.DestType != "OSS" && o.DestType != "NFS" {
		return errors.New("Invalid destination type")
	}
	if o.FullCrontab == "" {
		return errors.New("Full backup schedule is required, at least one of the full schedule must be specified")
	}
	if !checkCrontabSyntax(o.FullCrontab) {
		return errors.New("Invalid full backup schedule")
	}
	if o.IncrementalCrontab != "" && !checkCrontabSyntax(o.IncrementalCrontab) {
		return errors.New("Invalid incremental backup schedule")
	}
	return nil
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	o.AddBaseFlags(cmd)
	o.AddDaysFieldFlags(cmd)
	o.AddScheduleFlags(cmd)
	o.AddAccessFlags(cmd)
}

// AddBaseFlags adds the base flags for the create command
func (o *CreateOptions) AddBaseFlags(cmd *cobra.Command) {
	baseFlags := cmd.Flags()
	baseFlags.StringVar(&o.Name, FLAG_NAME, "", "The name of the ob tenant")
	baseFlags.StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of the ob tenant")
	baseFlags.StringVar(&o.DestType, FLAG_DEST_TYPE, DEFAULT_DEST_TYPE, "The destination type of the backup policy, currently support OSS or NFS")
	baseFlags.StringVar(&o.ArchivePath, FLAG_ARCHIVE_PATH, "", "The archive path of the backup policy")
	baseFlags.StringVar(&o.BakDataPath, FLAG_BAK_DATA_PATH, "", "The backup data path of the backup policy")
}

// AddDaysFieldFlags adds the days-field-related flags for the create command
func (o *CreateOptions) AddDaysFieldFlags(cmd *cobra.Command) {
	daysFieldFlags := pflag.NewFlagSet(FLAGSET_DAYS_FIELD, pflag.ContinueOnError)
	daysFieldFlags.IntVar(&o.JobKeepDays, FLAG_JOB_KEEP_DAYS, DEFAULT_JOB_KEEP_DAYS, "The days to keep the backup job")
	daysFieldFlags.IntVar(&o.RecoveryDays, FLAG_RECOVERY_DAYS, DEFAULT_RECOVERY_DAYS, "The days to keep the recovery job")
	cmd.Flags().AddFlagSet(daysFieldFlags)
}

// AddScheduleFlags adds the schedule-related flags for the create command
func (o *CreateOptions) AddScheduleFlags(cmd *cobra.Command) {
	scheduleFlags := pflag.NewFlagSet(FLAGSET_SCHEDULE, pflag.ContinueOnError)
	scheduleFlags.StringVar(&o.IncrementalCrontab, FLAG_INCREMENTAL, "", "The incremental backup schedule, crontab format, e.g. 0 0 * * 1,2,3")
	scheduleFlags.StringVar(&o.FullCrontab, FLAG_FULL, "", "The full backup schedule, crontab format, e.g. 0 0 * * 4,5")
	cmd.Flags().AddFlagSet(scheduleFlags)
}

// AddAccessFlags adds the access-related flags for the create command
func (o *CreateOptions) AddAccessFlags(cmd *cobra.Command) {
	accessFlags := pflag.NewFlagSet(FLAGSET_ACCESS, pflag.ContinueOnError)
	accessFlags.StringVar(&o.OSSAccessID, FLAG_OSS_ACCESS_ID, "", "The OSS access id for OSS destination")
	accessFlags.StringVar(&o.OSSAccessKey, FLAG_OSS_ACCESS_KEY, "", "The OSS access key for OSS destination")
	accessFlags.StringVar(&o.BakEncryptionPassword, FLAG_BAK_ENCRYPTION_PASSWORD, "", "The backup encryption password")
	cmd.Flags().AddFlagSet(accessFlags)
}
