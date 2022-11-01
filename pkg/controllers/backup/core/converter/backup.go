/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package converter

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
)

func GenerateBackupSpec(obBackup cloudv1.Backup) cloudv1.BackupSpec {
	// 好像直接 return obBackup.Spec 就行了呃
	spec := cloudv1.BackupSpec{
		DestPath:      obBackup.Spec.DestPath,
		SourceCluster: obBackup.Spec.SourceCluster,
		Schedule:      obBackup.Spec.Schedule,
		Parameters:    obBackup.Spec.Parameters,
	}
	return spec
}
