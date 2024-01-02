package resource

import (
	"context"

	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListAllPods() (*corev1.PodList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}

func ListPods(namespace string) (*corev1.PodList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}
