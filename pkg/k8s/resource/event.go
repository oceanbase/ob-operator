package resource

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func ListAllEvents(listOptions *metav1.ListOptions) (*corev1.EventList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), *listOptions)
}

func ListEvents(namespace string, listOptions *metav1.ListOptions) (*corev1.EventList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Events(namespace).List(context.TODO(), *listOptions)
}
