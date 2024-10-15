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
	"strings"

	"github.com/oceanbase/ob-operator/internal/cli/generic"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

type UpdateOptions struct {
	generic.ResourceOption
	ScheduleType       string `json:"scheduleType" example:"Weekly"`
	JobKeepDays        int    `json:"jobKeepDays,omitempty" example:"5"`
	RecoveryDays       int    `json:"recoveryDays,omitempty" example:"3"`
	PieceIntervalDays  int    `json:"pieceIntervalDays,omitempty" example:"1"`
	IncrementalCrontab string `json:"incremental,omitempty"`
	FullCrontab        string `json:"full,omitempty"`
	// Description: HH:MM
	// Example: 04:00
	ScheduleTime string `json:"scheduleTime" example:"04:00"`
	Status       string `json:"status,omitempty"`
}

func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{}
}

func UpdateTenantBackupPolicy(ctx context.Context, o *UpdateOptions) error {
	nn := types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
	policy, err := clients.GetTenantBackupPolicy(ctx, nn)
	if err != nil {
		return err
	}
	if o.JobKeepDays != 0 {
		policy.Spec.JobKeepWindow = numberToDay(o.JobKeepDays)
	}
	if o.RecoveryDays != 0 {
		policy.Spec.DataClean.RecoveryWindow = numberToDay(o.RecoveryDays)
	}
	if o.PieceIntervalDays != 0 {
		policy.Spec.LogArchive.SwitchPieceInterval = numberToDay(o.PieceIntervalDays)
	}
	if strings.ToUpper(o.Status) == "PAUSED" {
		policy.Spec.Suspend = true
	} else if strings.ToUpper(o.Status) == "RUNNING" {
		policy.Spec.Suspend = false
	}
	if o.FullCrontab != "" || o.IncrementalCrontab != "" {
		policy.Spec.DataBackup.IncrementalCrontab = o.IncrementalCrontab
		policy.Spec.DataBackup.FullCrontab = o.FullCrontab
	}
	if _, err := clients.UpdateTenantBackupPolicy(ctx, policy); err != nil {
		return err
	}
	return nil
}

func (o *UpdateOptions) Validate() error {
	if o.JobKeepDays == 0 {
		return errors.New("jobKeepDays can not be zero")
	}
	if o.RecoveryDays == 0 {
		return errors.New("recoveryDays can not be zero")
	}
	if o.PieceIntervalDays == 0 {
		return errors.New("pieceIntervalDays can not be zero")
	}
	return nil
}

// AddFlags add basic flags for tenant management
func (o *UpdateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Name, FLAG_NAME, "", "The name of the ob tenant")
	cmd.Flags().StringVar(&o.Namespace, FLAG_NAMESPACE, DEFAULT_NAMESPACE, "The namespace of the ob tenant")
	cmd.Flags().IntVar(&o.JobKeepDays, FLAG_JOB_KEEP_DAYS, DEFAULT_JOB_KEEP_DAYS, "The number of days to keep the backup job")
	cmd.Flags().IntVar(&o.RecoveryDays, FLAG_RECOVERY_DAYS, DEFAULT_RECOVERY_DAYS, "The number of days to keep the backup recovery")
	cmd.Flags().IntVar(&o.PieceIntervalDays, FLAG_PIECE_INTERVAL_DAYS, DEFAULT_PIECE_INTERVAL_DAYS, "The number of days to switch the backup piece")
}
