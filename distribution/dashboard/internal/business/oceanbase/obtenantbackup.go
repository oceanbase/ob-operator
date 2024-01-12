package oceanbase

import (
	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/oceanbase-dashboard/internal/business/constant"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
)

func buildBackupPolicyApiType(nn types.NamespacedName, obcluster string, p *param.CreateBackupPolicy) *v1alpha1.OBTenantBackupPolicy {
	policy := &v1alpha1.OBTenantBackupPolicy{}
	policy.Name = nn.Name + "-backup-policy"
	policy.Namespace = nn.Namespace
	policy.Spec = v1alpha1.OBTenantBackupPolicySpec{
		ObClusterName: obcluster,
		TenantCRName:  nn.Name,
		JobKeepWindow: p.JobKeepWindow,
		LogArchive: v1alpha1.LogArchiveConfig{
			Destination: apitypes.BackupDestination{
				Path:            p.ArchiveSource,
				Type:            apitypes.BackupDestType(p.DestType),
				OSSAccessSecret: "",
			},
			SwitchPieceInterval: "1d",
		},
		DataBackup: v1alpha1.DataBackupConfig{
			Destination: apitypes.BackupDestination{
				Path:            p.BakDataSource,
				Type:            apitypes.BackupDestType(p.DestType),
				OSSAccessSecret: "",
			},
			FullCrontab:        "",
			IncrementalCrontab: "",
			EncryptionSecret:   "",
		},
		DataClean: v1alpha1.CleanPolicy{
			RecoveryWindow: p.RecoveryWindow,
		},
	}
	if p.DestType == "OSS" && p.OSSAccessID != "" && p.OSSAccessKey != "" {
		ossSecretName := nn.Name + "-backup-oss-secret-" + rand.String(6)
		policy.Spec.LogArchive.Destination.OSSAccessSecret = ossSecretName
		policy.Spec.DataBackup.Destination.OSSAccessSecret = ossSecretName
	}
	if p.BakEncryptionPassword != "" {
		encryptionSecretName := nn.Name + "-backup-encryption-secret-" + rand.String(6)
		policy.Spec.DataBackup.EncryptionSecret = encryptionSecretName
	}
	// TODO:
	if p.ScheduleType == "Weekly" {
		policy.Spec.DataBackup.FullCrontab = "0 0 * * 0"
		policy.Spec.DataBackup.IncrementalCrontab = "0 0 * * 1-6"
	} else if p.ScheduleType == "Monthly" {
		policy.Spec.DataBackup.FullCrontab = "0 0 1 * *"
		policy.Spec.DataBackup.IncrementalCrontab = "0 0 2-31 * *"
	}
	return policy
}

func buildBackupPolicyModelType(p *v1alpha1.OBTenantBackupPolicy) *response.BackupPolicy {
	res := &response.BackupPolicy{
		BackupPolicyBase: param.BackupPolicyBase{
			DestType:      param.BackupDestType(p.Spec.DataBackup.Destination.Type),
			ArchiveSource: p.Spec.LogArchive.Destination.Path,
			BakDataSource: p.Spec.DataBackup.Destination.Path,
			// TODO:
			ScheduleType:   "",
			ScheduleDates:  []param.ScheduleDate{},
			JobKeepWindow:  p.Spec.JobKeepWindow,
			RecoveryWindow: p.Spec.DataClean.RecoveryWindow,
			PieceInterval:  p.Spec.LogArchive.SwitchPieceInterval,
		},
		TenantName:          p.Spec.TenantCRName,
		Name:                p.Name,
		Namespace:           p.Namespace,
		Status:              string(p.Status.Status),
		OSSAccessSecret:     p.Spec.LogArchive.Destination.OSSAccessSecret,
		BakEncryptionSecret: p.Spec.DataBackup.EncryptionSecret,
	}
	return res
}

func buildBackupJobModelType(p *v1alpha1.OBTenantBackup) *response.BackupJob {
	if p == nil {
		return nil
	}
	res := &response.BackupJob{
		Name:             p.Name,
		Namespace:        p.Name,
		Type:             string(p.Spec.Type),
		TenantName:       p.Spec.TenantName,
		BackupPolicyName: "",
		Path:             p.Spec.Path,
		StartTime:        p.Status.StartedAt,
		EndTime:          p.Status.EndedAt,
		Status:           string(p.Status.Status),
		StatusInDatabase: "",
		EncryptionSecret: p.Spec.EncryptionSecret,
	}
	if p.Annotations != nil {
		if policyName, exist := p.Annotations[oceanbaseconst.LabelRefBackupPolicy]; exist {
			res.BackupPolicyName = policyName
		}
	}
	switch p.Spec.Type {
	case apiconst.BackupJobTypeFull, apiconst.BackupJobTypeIncr:
		res.StatusInDatabase = p.Status.BackupJob.Status
	case apiconst.BackupJobTypeArchive:
		res.StatusInDatabase = p.Status.ArchiveLogJob.Status
	case apiconst.BackupJobTypeClean:
		res.StatusInDatabase = p.Status.DataCleanJob.Status
	}
	return res
}

func GetTenantBackupPolicy(nn types.NamespacedName) (*response.BackupPolicy, error) {
	_, err := oceanbase.GetOBTenant(nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, NewOBError(ErrorTypeNotFound, "Tenant not found")
		}
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	policy, err := oceanbase.GetTenantBackupPolicy(nn)
	if err != nil {
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	return buildBackupPolicyModelType(policy), nil
}

func CreateTenantBackupPolicy(nn types.NamespacedName, p *param.CreateBackupPolicy) (*response.BackupPolicy, error) {
	tenant, err := oceanbase.GetOBTenant(nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, NewOBError(ErrorTypeNotFound, "Tenant not found")
		}
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	if tenant.Status.Status != "running" {
		return nil, NewOBError(ErrorTypeBadRequest, "Tenant is not running")
	}
	backupPolicy := buildBackupPolicyApiType(nn, tenant.Spec.ClusterName, p)
	policy, err := oceanbase.CreateTenantBackupPolicy(backupPolicy)
	if err != nil {
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	return buildBackupPolicyModelType(policy), nil
}

func UpdateTenantBackupPolicy(nn types.NamespacedName, p *param.UpdateBackupPolicy) (*response.BackupPolicy, error) {
	tenant, err := oceanbase.GetOBTenant(nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, NewOBError(ErrorTypeNotFound, "Tenant not found")
		}
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	if tenant.Status.Status != "running" {
		return nil, NewOBError(ErrorTypeBadRequest, "Tenant is not running")
	}
	policy, err := oceanbase.GetTenantBackupPolicy(nn)
	if err != nil {
		return nil, NewOBError(ErrorTypeBadRequest, err.Error())
	}
	policy.Spec.JobKeepWindow = p.JobKeepWindow
	policy.Spec.DataClean.RecoveryWindow = p.RecoveryWindow
	policy.Spec.LogArchive.SwitchPieceInterval = p.PieceInterval
	if p.Status == "Paused" {
		policy.Spec.Suspend = true
	}
	if p.Status == "Running" {
		policy.Spec.Suspend = false
	}
	np, err := oceanbase.UpdateTenantBackupPolicy(policy)
	if err != nil {
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	return buildBackupPolicyModelType(np), nil
}

func DeleteTenantBackupPolicy(nn types.NamespacedName) error {
	return oceanbase.DeleteTenantBackupPolicy(nn)
}

func ListBackupJobs(nn types.NamespacedName, jobType string, limit int) ([]*response.BackupJob, error) {
	policy, err := oceanbase.GetTenantBackupPolicy(nn)
	if err != nil {
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	listOption := metav1.ListOptions{}
	if jobType != "" && jobType != "ALL" {
		listOption.LabelSelector = oceanbaseconst.LabelRefBackupPolicy + "=" + policy.Name + "," + oceanbaseconst.LabelBackupType + "=" + jobType
	} else {
		listOption.LabelSelector = oceanbaseconst.LabelRefBackupPolicy + "=" + policy.Name
	}
	listOption.Limit = int64(limit)
	jobs, err := oceanbase.ListBackupJobs(listOption)
	if err != nil {
		return nil, NewOBError(ErrorTypeInternal, err.Error())
	}
	res := make([]*response.BackupJob, 0)
	for _, job := range jobs.Items {
		res = append(res, buildBackupJobModelType(&job))
	}
	return res, nil
}
