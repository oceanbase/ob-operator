package client

import (
	"context"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("K8s", func() {
	// Simulate a dynamic-style pod client that behave like a static-style pod client in clientset
	It("Test pod client", func() {
		client := GetClient()
		pods := &corev1.PodList{}
		podList, err := client.ClientSet.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		podClient := NewDynamicResourceClient[*corev1.Pod](corev1.SchemeGroupVersion.WithResource("pods"), "Pod")
		err = podClient.List(context.Background(), "", pods, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())

		Expect(len(podList.Items)).Should(Equal(len(pods.Items)))
		for i := range podList.Items {
			Expect(podList.Items[i].Name).Should(Equal(pods.Items[i].Name))
		}

		randomIdx := rand.Intn(len(podList.Items))
		targetPod := podList.Items[randomIdx]
		pod, err := podClient.Get(context.Background(), targetPod.Namespace, targetPod.Name, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(targetPod.Name).Should(Equal(pod.Name))
	})
})
