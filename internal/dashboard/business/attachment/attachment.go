/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package attachment

import (
	"fmt"
	"os"

	"github.com/oceanbase/ob-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/util/rand"
)

func GetAttachment(id string) (string, error) {
	sharedMountPath := os.Getenv("SHARED_VOLUME_MOUNT_PATH")
	attachmentDir := fmt.Sprintf("%s/%s", sharedMountPath, id)
	// Create a temporary file in the shared mount path to avoid race conditions and ephemeral storage issues
	zipFile := fmt.Sprintf("%s/%s-%s.zip", sharedMountPath, id, rand.String(6))

	if err := utils.Zip(attachmentDir, zipFile); err != nil {
		return "", err
	}

	return zipFile, nil
}
