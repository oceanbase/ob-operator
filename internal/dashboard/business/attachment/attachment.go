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
