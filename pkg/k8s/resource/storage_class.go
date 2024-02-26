package resource

import (
	"context"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func ListStorageClasses() (*storagev1.StorageClassList, error) {
	client := client.GetClient()
	return client.ClientSet.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}
