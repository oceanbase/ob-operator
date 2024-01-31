package client

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Simulate a dynamic-style pod client that behave like a static-style pod client in clientset
func TestPodClient(t *testing.T) {
	client := GetClient()
	pods := &corev1.PodList{}
	podList, err := client.ClientSet.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	assert.Nil(t, err)
	podClient := NewDynamicResourceClient[*corev1.Pod](corev1.SchemeGroupVersion.WithResource("pods"), "Pod")
	err = podClient.List(context.Background(), "", pods, metav1.ListOptions{})
	assert.Nil(t, err)

	assert.Equal(t, len(podList.Items), len(pods.Items))
	for i := range podList.Items {
		assert.Equal(t, podList.Items[i].Name, pods.Items[i].Name)
	}

	randomIdx := rand.Intn(len(podList.Items))
	targetPod := podList.Items[randomIdx]
	pod, err := podClient.Get(context.Background(), targetPod.Namespace, targetPod.Name, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, targetPod.Name, pod.Name)
}
