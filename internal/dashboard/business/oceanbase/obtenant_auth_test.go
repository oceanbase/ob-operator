/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type authenticationTenantClient struct {
	client.K8sResourceClient[*v1alpha1.OBTenant]
	listCalls int
}

func (c *authenticationTenantClient) List(_ context.Context, _ string, list runtimeclient.ObjectList, _ metav1.ListOptions) error {
	c.listCalls++
	list.(*v1alpha1.OBTenantList).Items = nil
	return nil
}

func TestListAllOBTenantsRequiresIdentity(t *testing.T) {
	for _, tc := range []struct {
		name     string
		username interface{}
	}{
		{"missing", nil},
		{"empty", ""},
		{"wrong type", 123},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &authenticationTenantClient{}
			previous := clients.TenantClient
			clients.TenantClient = fake
			t.Cleanup(func() { clients.TenantClient = previous })
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tc.username != nil {
				ctx.Set("username", tc.username)
			}
			require.NotPanics(t, func() {
				tenants, err := ListAllOBTenants(ctx, "test", metav1.ListOptions{})
				require.Nil(t, tenants)
				var authErr oberr.ObError
				require.ErrorAs(t, err, &authErr)
				require.Equal(t, http.StatusUnauthorized, authErr.Status())
			})
			require.Zero(t, fake.listCalls, "reject missing identity before reading tenant data")
		})
	}
}

func TestListAllOBTenantsWithIdentity(t *testing.T) {
	fake := &authenticationTenantClient{}
	previous := clients.TenantClient
	clients.TenantClient = fake
	t.Cleanup(func() { clients.TenantClient = previous })
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set("username", "test-user")
	tenants, err := ListAllOBTenants(ctx, "test", metav1.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, tenants)
	require.Equal(t, 1, fake.listCalls)
}
