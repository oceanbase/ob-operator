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

package response

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

type BackupPolicy struct {
	param.BackupPolicyBase `json:",inline"`

	UID                 string `json:"uid" binding:"required"`
	TenantName          string `json:"tenantName" binding:"required"`
	Name                string `json:"name" binding:"required"`
	Namespace           string `json:"namespace" binding:"required"`
	Status              string `json:"status" binding:"required"`
	OSSAccessSecret     string `json:"ossAccessSecret,omitempty"`
	BakEncryptionSecret string `json:"bakEncryptionSecret,omitempty"`

	CreateTime string     `json:"createTime" binding:"required"`
	Events     []K8sEvent `json:"events" binding:"required"`
}

type BackupJob struct {
	UID       string `json:"uid" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	// Enum: FULL, INCR, ARCHIVE, CLEAN
	Type             string `json:"type" binding:"required"`
	TenantName       string `json:"tenantName" binding:"required"`
	BackupPolicyName string `json:"backupPolicyName" binding:"required"`
	Path             string `json:"path" binding:"required"`      // Empty for Clean job
	StartTime        string `json:"startTime" binding:"required"` // Start time of the backup job, StartScnDisplay for ARCHIVE job
	EndTime          string `json:"endTime"`                      // End time of the backup job, empty for ARCHIVE job
	Status           string `json:"status" binding:"required"`
	StatusInDatabase string `json:"statusInDatabase" binding:"required"`
	EncryptionSecret string `json:"encryptionSecret,omitempty"`
}
