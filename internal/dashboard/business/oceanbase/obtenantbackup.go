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
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconst "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/internal/oceanbase"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func numberToDay(n int) string {
	return fmt.Sprintf("%dd", n)
}

func dayToNumber(day string) int {
	if !strings.HasSuffix(day, "d") {
		return 0
	}
	n, err := strconv.Atoi(day[:len(day)-1])
	if err != nil {
		return 0
	}
	return n
}

func setScheduleDatesToPolicy(policy *v1alpha1.OBTenantBackupPolicy, p param.ScheduleBase) {
	hourMinutes := strings.Split(p.ScheduleTime, ":")
	crontabParts := fmt.Sprintf("%s %s", hourMinutes[1], hourMinutes[0])

	if p.ScheduleType == "Weekly" {
		crontabParts += " * *"
		fullCrontabWeekdays := make([]string, 0)
		incrementalCrontabWeekdays := make([]string, 0)
		for _, date := range p.ScheduleDates {
			if date.BackupType == "Full" {
				fullCrontabWeekdays = append(fullCrontabWeekdays, fmt.Sprint(date.Day%7))
			} else if date.BackupType == "Incremental" {
				incrementalCrontabWeekdays = append(incrementalCrontabWeekdays, fmt.Sprint(date.Day%7))
			}
		}
		policy.Spec.DataBackup.FullCrontab = crontabParts + " " + strings.Join(fullCrontabWeekdays, ",")
		if len(incrementalCrontabWeekdays) > 0 {
			policy.Spec.DataBackup.IncrementalCrontab = crontabParts + " " + strings.Join(incrementalCrontabWeekdays, ",")
		} else {
			policy.Spec.DataBackup.IncrementalCrontab = crontabParts + " *"
		}
	} else if p.ScheduleType == "Monthly" {
		fullCrontabMonthdays := make([]string, 0)
		incrementalCrontabMonthdays := make([]string, 0)
		for _, date := range p.ScheduleDates {
			if date.BackupType == "Full" {
				fullCrontabMonthdays = append(fullCrontabMonthdays, fmt.Sprint(date.Day))
			} else if date.BackupType == "Incremental" {
				incrementalCrontabMonthdays = append(incrementalCrontabMonthdays, fmt.Sprint(date.Day))
			}
		}
		policy.Spec.DataBackup.FullCrontab = strings.Join([]string{crontabParts, strings.Join(fullCrontabMonthdays, ","), "* *"}, " ")
		if len(incrementalCrontabMonthdays) > 0 {
			policy.Spec.DataBackup.IncrementalCrontab = strings.Join([]string{crontabParts, strings.Join(incrementalCrontabMonthdays, ","), "* *"}, " ")
		} else {
			policy.Spec.DataBackup.IncrementalCrontab = strings.Join([]string{crontabParts, "*", "* *"}, " ")
		}
	}
}

func getScheduleDatesFromPolicy(p *v1alpha1.OBTenantBackupPolicy) param.ScheduleBase {
	res := param.ScheduleBase{}
	fullParts := strings.Split(p.Spec.DataBackup.FullCrontab, " ")
	incrementalParts := strings.Split(p.Spec.DataBackup.IncrementalCrontab, " ")
	res.ScheduleTime = fmt.Sprintf("%s:%s", fullParts[1], fullParts[0])
	var fullDays, incrementalDays []string
	var processDay func(day int) int

	// Ends with "*", means the type is Monthly
	if strings.HasSuffix(p.Spec.DataBackup.FullCrontab, "*") {
		res.ScheduleType = "Monthly"
		fullDays = strings.Split(fullParts[2], ",")
		incrementalDays = strings.Split(incrementalParts[2], ",")
		processDay = func(day int) int {
			return day
		}
	} else {
		res.ScheduleType = "Weekly"
		fullDays = strings.Split(fullParts[4], ",")
		incrementalDays = strings.Split(incrementalParts[4], ",")
		// Crontab use 0-6 to represent Sunday to Saturday, but we use 1-7
		processDay = func(day int) int {
			if day == 0 {
				return 7
			}
			return day
		}
	}
	var i, j int
	for i < len(fullDays) && j < len(incrementalDays) {
		fullDay, _ := strconv.Atoi(fullDays[i])
		if incrementalDays[j] == "*" {
			j = len(incrementalDays)
			break
		}
		incrementalDay, _ := strconv.Atoi(incrementalDays[j])
		// It should not happen, but just in case
		if fullDay == incrementalDay {
			res.ScheduleDates = append(res.ScheduleDates, param.ScheduleDate{
				Day:        processDay(fullDay),
				BackupType: "Full",
			})
			i++
			j++
		} else if fullDay < incrementalDay {
			res.ScheduleDates = append(res.ScheduleDates, param.ScheduleDate{
				Day:        processDay(fullDay),
				BackupType: "Full",
			})
			i++
		} else {
			res.ScheduleDates = append(res.ScheduleDates, param.ScheduleDate{
				Day:        processDay(incrementalDay),
				BackupType: "Incremental",
			})
			j++
		}
	}
	for i < len(fullDays) {
		fullDay, _ := strconv.Atoi(fullDays[i])
		res.ScheduleDates = append(res.ScheduleDates, param.ScheduleDate{
			Day:        processDay(fullDay),
			BackupType: "Full",
		})
		i++
	}
	for j < len(incrementalDays) {
		incrementalDay, _ := strconv.Atoi(incrementalDays[j])
		res.ScheduleDates = append(res.ScheduleDates, param.ScheduleDate{
			Day:        processDay(incrementalDay),
			BackupType: "Incremental",
		})
		j++
	}
	return res
}

func buildBackupPolicyApiType(nn types.NamespacedName, obcluster string, p *param.CreateBackupPolicy) *v1alpha1.OBTenantBackupPolicy {
	policy := &v1alpha1.OBTenantBackupPolicy{}
	policy.Name = nn.Name + "-backup-policy"
	policy.Namespace = nn.Namespace
	policy.Spec = v1alpha1.OBTenantBackupPolicySpec{
		ObClusterName: obcluster,
		TenantCRName:  nn.Name,
		JobKeepWindow: numberToDay(p.JobKeepDays),
		LogArchive: v1alpha1.LogArchiveConfig{
			Destination: apitypes.BackupDestination{
				Path:            p.ArchivePath,
				Type:            apitypes.BackupDestType(p.DestType),
				OSSAccessSecret: "",
			},
			SwitchPieceInterval: "1d",
		},
		DataBackup: v1alpha1.DataBackupConfig{
			Destination: apitypes.BackupDestination{
				Path:            p.BakDataPath,
				Type:            apitypes.BackupDestType(p.DestType),
				OSSAccessSecret: "",
			},
			FullCrontab:        "",
			IncrementalCrontab: "",
			EncryptionSecret:   "",
		},
		DataClean: v1alpha1.CleanPolicy{
			RecoveryWindow: numberToDay(p.RecoveryDays),
		},
	}

	setScheduleDatesToPolicy(policy, p.ScheduleBase)

	return policy
}

func buildBackupPolicyModelType(p *v1alpha1.OBTenantBackupPolicy) *response.BackupPolicy {
	res := &response.BackupPolicy{
		BackupPolicyBase: param.BackupPolicyBase{
			DestType:    param.BackupDestType(p.Spec.DataBackup.Destination.Type),
			ArchivePath: p.Spec.LogArchive.Destination.Path,
			BakDataPath: p.Spec.DataBackup.Destination.Path,
			ScheduleBase: param.ScheduleBase{
				ScheduleType:  "",
				ScheduleTime:  "",
				ScheduleDates: []param.ScheduleDate{},
			},
			DaysFieldBase: param.DaysFieldBase{
				JobKeepDays:       dayToNumber(p.Spec.JobKeepWindow),
				RecoveryDays:      dayToNumber(p.Spec.DataClean.RecoveryWindow),
				PieceIntervalDays: dayToNumber(p.Spec.LogArchive.SwitchPieceInterval),
			},
		},
		TenantName:          p.Spec.TenantCRName,
		Name:                p.Name,
		Namespace:           p.Namespace,
		Status:              string(p.Status.Status),
		OSSAccessSecret:     p.Spec.LogArchive.Destination.OSSAccessSecret,
		BakEncryptionSecret: p.Spec.DataBackup.EncryptionSecret,
	}
	res.ScheduleBase = getScheduleDatesFromPolicy(p)
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
		if p.Status.BackupJob != nil {
			res.StatusInDatabase = p.Status.BackupJob.Status
		}
	case apiconst.BackupJobTypeArchive:
		if p.Status.ArchiveLogJob != nil {
			res.StatusInDatabase = p.Status.ArchiveLogJob.Status
		}
	case apiconst.BackupJobTypeClean:
		if p.Status.DataCleanJob != nil {
			res.StatusInDatabase = p.Status.DataCleanJob.Status
		}
	}
	return res
}

func GetTenantBackupPolicy(ctx context.Context, nn types.NamespacedName) (*response.BackupPolicy, error) {
	_, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, oberr.NewNotFound("Tenant not found")
		}
		return nil, oberr.NewInternal(err.Error())
	}
	policy, err := oceanbase.GetTenantBackupPolicy(ctx, nn)
	if err != nil {
		return nil, oberr.NewInternal(err.Error())
	}
	if policy == nil {
		return nil, nil
	}
	return buildBackupPolicyModelType(policy), nil
}

func CreateTenantBackupPolicy(ctx context.Context, nn types.NamespacedName, p *param.CreateBackupPolicy) (*response.BackupPolicy, error) {
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, oberr.NewNotFound("Tenant not found")
		}
		return nil, oberr.NewInternal(err.Error())
	}
	if tenant.Status.Status != "running" {
		return nil, oberr.NewBadRequest("Tenant is not running")
	}
	if p.JobKeepDays == 0 {
		p.JobKeepDays = 7
	}
	if p.RecoveryDays == 0 {
		p.RecoveryDays = 30
	}
	if p.PieceIntervalDays == 0 {
		p.PieceIntervalDays = 1
	}
	backupPolicy := buildBackupPolicyApiType(nn, tenant.Spec.ClusterName, p)

	if p.DestType == "OSS" && p.OSSAccessID != "" && p.OSSAccessKey != "" {
		ossSecretName := nn.Name + "-backup-oss-secret-" + rand.String(6)
		backupPolicy.Spec.LogArchive.Destination.OSSAccessSecret = ossSecretName
		backupPolicy.Spec.DataBackup.Destination.OSSAccessSecret = ossSecretName
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ossSecretName,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"accessId":  p.OSSAccessID,
				"accessKey": p.OSSAccessKey,
			},
		}
		_, err := client.GetClient().ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, oberr.NewInternal(err.Error())
		}
	}
	if p.BakEncryptionPassword != "" {
		encryptionSecretName := nn.Name + "-backup-encryption-secret-" + rand.String(6)
		backupPolicy.Spec.DataBackup.EncryptionSecret = encryptionSecretName
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      encryptionSecretName,
				Namespace: nn.Namespace,
			},
			StringData: map[string]string{
				"password": p.BakEncryptionPassword,
			},
		}
		_, err := client.GetClient().ClientSet.CoreV1().Secrets(nn.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, oberr.NewInternal(err.Error())
		}
	}

	policy, err := oceanbase.CreateTenantBackupPolicy(ctx, backupPolicy)
	if err != nil {
		return nil, oberr.NewInternal(err.Error())
	}
	return buildBackupPolicyModelType(policy), nil
}

func UpdateTenantBackupPolicy(ctx context.Context, nn types.NamespacedName, p *param.UpdateBackupPolicy) (*response.BackupPolicy, error) {
	tenant, err := oceanbase.GetOBTenant(ctx, nn)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, oberr.NewNotFound("Tenant not found")
		}
		return nil, oberr.NewInternal(err.Error())
	}
	if tenant.Status.Status != "running" {
		return nil, oberr.NewBadRequest("Tenant is not running")
	}
	policy, err := oceanbase.GetTenantBackupPolicy(ctx, nn)
	if err != nil {
		return nil, oberr.NewBadRequest(err.Error())
	}
	if p.JobKeepDays != 0 {
		policy.Spec.JobKeepWindow = numberToDay(p.JobKeepDays)
	}
	if p.RecoveryDays != 0 {
		policy.Spec.DataClean.RecoveryWindow = numberToDay(p.RecoveryDays)
	}
	if p.PieceIntervalDays != 0 {
		policy.Spec.LogArchive.SwitchPieceInterval = numberToDay(p.PieceIntervalDays)
	}

	if strings.ToUpper(p.Status) == "PAUSED" {
		policy.Spec.Suspend = true
	} else if strings.ToUpper(p.Status) == "RUNNING" {
		policy.Spec.Suspend = false
	}

	schedule := p.ScheduleBase
	if schedule.ScheduleDates != nil || schedule.ScheduleTime != "" || schedule.ScheduleType != "" {
		overlaySchedule := getScheduleDatesFromPolicy(policy)
		if schedule.ScheduleType != "" {
			overlaySchedule.ScheduleType = schedule.ScheduleType
		}
		if schedule.ScheduleTime != "" {
			overlaySchedule.ScheduleTime = schedule.ScheduleTime
		}
		if schedule.ScheduleDates != nil {
			overlaySchedule.ScheduleDates = schedule.ScheduleDates
		}
		setScheduleDatesToPolicy(policy, overlaySchedule)
	}

	np, err := oceanbase.UpdateTenantBackupPolicy(ctx, policy)
	if err != nil {
		return nil, oberr.NewInternal(err.Error())
	}
	return buildBackupPolicyModelType(np), nil
}

func DeleteTenantBackupPolicy(ctx context.Context, nn types.NamespacedName) error {
	policy, err := oceanbase.GetTenantBackupPolicy(ctx, nn)
	if err != nil {
		return oberr.NewBadRequest(err.Error())
	}
	return oceanbase.DeleteTenantBackupPolicy(ctx, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace})
}

func ListBackupJobs(ctx context.Context, nn types.NamespacedName, jobType string, limit int) ([]*response.BackupJob, error) {
	policy, err := oceanbase.GetTenantBackupPolicy(ctx, nn)
	if err != nil {
		return nil, oberr.NewInternal(err.Error())
	}
	if policy == nil {
		return nil, nil
	}
	listOption := metav1.ListOptions{}
	if jobType != "" && jobType != "ALL" {
		listOption.LabelSelector = oceanbaseconst.LabelRefBackupPolicy + "=" + policy.Name + "," + oceanbaseconst.LabelBackupType + "=" + jobType
	} else {
		listOption.LabelSelector = oceanbaseconst.LabelRefBackupPolicy + "=" + policy.Name
	}
	listOption.Limit = int64(limit)
	jobs, err := oceanbase.ListBackupJobs(ctx, listOption)
	if err != nil {
		return nil, oberr.NewInternal(err.Error())
	}
	res := make([]*response.BackupJob, 0)
	for _, job := range jobs.Items {
		res = append(res, buildBackupJobModelType(&job))
	}
	return res, nil
}
