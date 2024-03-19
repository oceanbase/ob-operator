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

	TenantName          string `json:"tenantName"`
	Name                string `json:"name"`
	Namespace           string `json:"namespace"`
	Status              string `json:"status"`
	OSSAccessSecret     string `json:"ossAccessSecret,omitempty"`
	BakEncryptionSecret string `json:"bakEncryptionSecret,omitempty"`

	CreateTime string     `json:"createTime"`
	Events     []K8sEvent `json:"events"`
}

type BackupJob struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// Enum: FULL, INCR, ARCHIVE, CLEAN
	Type             string `json:"type"`
	TenantName       string `json:"tenantName"`
	BackupPolicyName string `json:"backupPolicyName"`
	Path             string `json:"path"`      // Empty for Clean job
	StartTime        string `json:"startTime"` // Start time of the backup job, StartScnDisplay for ARCHIVE job
	EndTime          string `json:"endTime"`   // End time of the backup job, empty for ARCHIVE job
	Status           string `json:"status"`
	StatusInDatabase string `json:"statusInDatabase"`
	EncryptionSecret string `json:"encryptionSecret,omitempty"`
}
