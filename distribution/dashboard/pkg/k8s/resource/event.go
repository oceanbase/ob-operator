package resource

import (
	"context"

	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListAllEvents() (*corev1.EventList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}

func ListEvents(namespace string) (*corev1.EventList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}
