package oceanbase

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Oceanbase", func() {
	It("Test OBClusterClient", func() {
		clusterList := &v1alpha1.OBClusterList{}
		err := ClusterClient.List(context.Background(), "", clusterList, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(clusterList.Items).ShouldNot(BeNil())
	})

	It("Test OBTenantClient", func() {
		tenantList := &v1alpha1.OBTenantList{}
		err := TenantClient.List(context.Background(), "", tenantList, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test OBBackupPolicyClient", func() {
		policies := v1alpha1.OBTenantBackupPolicyList{}
		err := BackupPolicyClient.List(context.Background(), "", &policies, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test OBBackupJobClient", func() {
		jobs := v1alpha1.OBTenantBackupList{}
		err := BackupJobClient.List(context.Background(), "", &jobs, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test OBTenantOperationClient", func() {
		operations := v1alpha1.OBTenantOperationList{}
		err := OperationClient.List(context.Background(), "", &operations, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test OBServerClinet", func() {
		servers := v1alpha1.OBServerList{}
		err := ServerClient.List(context.Background(), "", &servers, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Test OBZoneClient", func() {
		zones := v1alpha1.OBZoneList{}
		err := ZoneClient.List(context.Background(), "", &zones, metav1.ListOptions{})
		Expect(err).ShouldNot(HaveOccurred())
	})
})
