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

import "github.com/oceanbase/ob-operator/internal/telemetry/models"

type StatisticData struct {
	Version        string                  `json:"version"`
	Clusters       []models.OBCluster      `json:"clusters"`
	Zones          []models.OBZone         `json:"zones"`
	Servers        []models.OBServer       `json:"servers"`
	Tenants        []models.OBTenant       `json:"tenants"`
	BackupPolicies []models.OBBackupPolicy `json:"backupPolicies"`
	WarningEvents  []models.K8sEvent       `json:"warningEvents"`
}

type StatisticDataResponse struct {
	Component string         `json:"component"`
	Time      string         `json:"time"`
	Content   *StatisticData `json:"content"`
}
