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

package utils

import (
	"context"
	"path"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

// BackupDestinationPath resolves the OceanBase storage URI used for both backup and restore.
func BackupDestinationPath(ctx context.Context, reader client.Reader, namespace string, dest *apitypes.BackupDestination) (string, error) {
	if dest == nil {
		return "", nil
	}
	if dest.Type == "" || dest.Type == constants.BackupDestTypeNFS {
		return "file://" + path.Join(oceanbaseconst.BackupPath, dest.Path), nil
	}
	if dest.OSSAccessSecret == "" {
		return "", nil
	}
	var secret corev1.Secret
	if err := reader.Get(ctx, client.ObjectKey{Namespace: namespace, Name: dest.OSSAccessSecret}, &secret); err != nil {
		return "", err
	}
	uri := dest.Path + "&access_id=" + string(secret.Data["accessId"]) + "&access_key=" + string(secret.Data["accessKey"])
	switch dest.Type {
	case constants.BackupDestTypeCOS:
		uri += "&appid=" + string(secret.Data["appId"])
	case constants.BackupDestTypeS3:
		uri += "&s3_region=" + string(secret.Data["s3Region"])
	case constants.BackupDestTypeS3Compatible:
		if model, ok := secret.Data["addressingModel"]; ok {
			uri += "&addressing_model=" + string(model)
		}
	}
	return uri, nil
}
