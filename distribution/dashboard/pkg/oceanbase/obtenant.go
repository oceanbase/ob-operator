package oceanbase

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase/schema"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func CreateOBTenant(tenant *v1alpha1.OBTenant) (*v1alpha1.OBTenant, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tenant)
	if err != nil {
		logger.Info("Convert tenant to unstructured", "err", err)
		return nil, errors.Wrap(err, "Convert tenant to unstructured")
	}
	tenantUnstructured := &unstructured.Unstructured{Object: objMap}
	tenantUnstructured.SetGroupVersionKind(schema.OBTenantResKind)
	newTenant, err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(tenant.Namespace).Create(context.TODO(), tenantUnstructured, metav1.CreateOptions{})
	if err != nil {
		logger.Info("Create tenant", "err", err)
		return nil, errors.Wrap(err, "Create tenant")
	}
	t := &v1alpha1.OBTenant{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(newTenant.UnstructuredContent(), t)
	if err != nil {
		logger.Info("Convert unstructured tenant to typed", "err", err)
		return nil, errors.Wrap(err, "Convert unstructured tenant to typed")
	}
	return t, nil
}

func UpdateOBTenant(tenant *v1alpha1.OBTenant) (*v1alpha1.OBTenant, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "Convert tenant to unstructured")
	}
	unstructuredTenant := &unstructured.Unstructured{Object: objMap}
	newTenant, err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(tenant.Namespace).Update(context.TODO(), unstructuredTenant, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Update tenant")
	}
	t := &v1alpha1.OBTenant{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(newTenant.UnstructuredContent(), t)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured tenant to typed")
	}
	return t, nil
}

func ListAllOBTenants() (*v1alpha1.OBTenantList, error) {
	clt := client.GetClient()
	tenantList, err := clt.DynamicClient.Resource(schema.OBTenantRes).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Info("List all tenants", "err", err)
		return nil, errors.Wrap(err, "List all tenants")
	}
	list := &v1alpha1.OBTenantList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(tenantList.UnstructuredContent(), list)
	if err != nil {
		logger.Info("Convert tenant list", "err", err)
		return nil, errors.Wrap(err, "Convert tenant list")
	}
	return list, nil
}

func GetOBTenant(nn types.NamespacedName) (*v1alpha1.OBTenant, error) {
	clt := client.GetClient()
	tenant, err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(nn.Namespace).Get(context.TODO(), nn.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Get tenant")
	}
	t := &v1alpha1.OBTenant{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(tenant.UnstructuredContent(), t)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured tenant to typed")
	}
	return t, nil
}

func DeleteOBTenant(nn types.NamespacedName) error {
	clt := client.GetClient()
	err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(nn.Namespace).Delete(context.TODO(), nn.Name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "Delete tenant")
	}
	return nil
}
