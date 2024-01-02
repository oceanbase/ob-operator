package resource

import (
	"context"

	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	k8sconst "github.com/oceanbase/oceanbase-dashboard/pkg/k8s/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var timeout int64 = k8sconst.DefaultClientListTimeoutSeconds

func ListNodes() (*corev1.NodeList, error) {
	client := client.GetClient()
	return client.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &timeout,
	})
}
