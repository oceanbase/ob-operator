package oceanbase

import (
	"context"
	"testing"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOBClusterClient(t *testing.T) {
	clusterList := &v1alpha1.OBClusterList{}
	err := ClusterClient.List(context.Background(), "", clusterList, metav1.ListOptions{})
	assert.Nil(t, err)
	assert.NotEmpty(t, clusterList.Items)
}

func TestOBTenantClient(t *testing.T) {
	tenantList := &v1alpha1.OBTenantList{}
	err := TenantClient.List(context.Background(), "", tenantList, metav1.ListOptions{})
	assert.Nil(t, err)
}

func TestOBBackupPolicyClient(t *testing.T) {
	policies := v1alpha1.OBTenantBackupPolicyList{}
	err := BackupPolicyClient.List(context.Background(), "", &policies, metav1.ListOptions{})
	assert.Nil(t, err)
}

func TestOBBackupJobClient(t *testing.T) {
	jobs := v1alpha1.OBTenantBackupList{}
	err := BackupJobClient.List(context.Background(), "", &jobs, metav1.ListOptions{})
	assert.Nil(t, err)
}

func TestOBTenantOperationClient(t *testing.T) {
	operations := v1alpha1.OBTenantOperationList{}
	err := OperationClient.List(context.Background(), "", &operations, metav1.ListOptions{})
	assert.Nil(t, err)
}

func TestOBServerClinet(t *testing.T) {
	servers := v1alpha1.OBServerList{}
	err := ServerClient.List(context.Background(), "", &servers, metav1.ListOptions{})
	assert.Nil(t, err)
}

func TestOBZoneClient(t *testing.T) {
	zones := v1alpha1.OBZoneList{}
	err := ZoneClient.List(context.Background(), "", &zones, metav1.ListOptions{})
	assert.Nil(t, err)
}
