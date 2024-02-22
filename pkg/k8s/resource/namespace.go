package resource

import (
	"context"

	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListNamespaces() (*corev1.NamespaceList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}

func CreateNamespace(namespace string) error {
	namespaceObject := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	client := client.GetClient()
	_, err := client.ClientSet.CoreV1().Namespaces().Create(context.TODO(), &namespaceObject, metav1.CreateOptions{})
	return err
}
