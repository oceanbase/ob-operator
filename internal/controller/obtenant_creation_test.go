package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
)

var _ = Describe("Ordinary OBTenant creation", func() {
	It("initializes a valid status and persists ownership before the first database task", func() {
		ctx := context.Background()
		size := apitypes.StorageSpec{Size: resource.MustParse("1Gi")}
		cluster := &v1alpha1.OBCluster{ObjectMeta: metav1.ObjectMeta{GenerateName: "creation-cluster-", Namespace: "default"},
			Spec: v1alpha1.OBClusterSpec{ClusterName: "example", ClusterId: 1, Topology: []apitypes.OBZoneTopology{}, UserSecrets: &apitypes.OBUserSecrets{Root: "root-ref"},
				OBServerTemplate: &apitypes.OBServerTemplate{Image: "example.invalid/observer:example", Resource: &apitypes.ResourceSpec{Memory: resource.MustParse("1Gi")},
					Storage: &apitypes.OceanbaseStorageSpec{DataStorage: &size, LogStorage: &size}}}}
		Expect(k8sClient.Create(ctx, cluster)).To(Succeed())
		tenant := &v1alpha1.OBTenant{ObjectMeta: metav1.ObjectMeta{GenerateName: "creation-tenant-", Namespace: "default", OwnerReferences: []metav1.OwnerReference{
			{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBCluster", Name: cluster.Name, UID: cluster.UID}}},
			Spec: v1alpha1.OBTenantSpec{ClusterName: cluster.Name, TenantName: "example", UnitNumber: 1, Pools: []v1alpha1.ResourcePoolSpec{}}}
		Expect(k8sClient.Create(ctx, tenant)).To(Succeed())
		key := client.ObjectKeyFromObject(tenant)
		r := &OBTenantReconciler{Client: k8sClient, APIReader: k8sClient, Scheme: k8sClient.Scheme(), Recorder: record.NewFakeRecorder(10)}
		for j := 0; j < 2; j++ {
			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})
			Expect(err).NotTo(HaveOccurred())
		}
		var live v1alpha1.OBTenant
		Expect(k8sClient.Get(ctx, key, &live)).To(Succeed())
		Expect(live.Status.Status).To(Equal(tenantstatus.CreatingTenant))
		Expect(live.Status.OperationContext).To(BeNil())
		Expect(live.Status.Pools).NotTo(BeNil())
		Expect(live.Status.Creation).NotTo(BeNil())
		Expect(live.Status.Creation.TenantUID).To(Equal(tenant.UID))
		Expect(live.Status.Creation.ClusterUID).To(Equal(cluster.UID))
		Expect(live.Status.Creation.DatabaseCluster).To(Equal(cluster.Spec.ClusterName))
		Expect(live.Status.Creation.TenantName).To(Equal(tenant.Spec.TenantName))
	})
})
