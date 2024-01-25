package oceanbase

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/oceanbase-dashboard/internal/business/constant"
	"github.com/oceanbase/oceanbase-dashboard/pkg/k8s/client"
	"github.com/oceanbase/oceanbase-dashboard/pkg/oceanbase/schema"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func CreateOBTenant(ctx context.Context, tenant *v1alpha1.OBTenant) (*v1alpha1.OBTenant, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tenant)
	if err != nil {
		logger.Info("Convert tenant to unstructured", "err", err)
		return nil, errors.Wrap(err, "Convert tenant to unstructured")
	}
	tenantUnstructured := &unstructured.Unstructured{Object: objMap}
	tenantUnstructured.SetGroupVersionKind(schema.OBTenantResKind)
	newTenant, err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(tenant.Namespace).Create(ctx, tenantUnstructured, metav1.CreateOptions{})
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

func UpdateOBTenant(ctx context.Context, tenant *v1alpha1.OBTenant) (*v1alpha1.OBTenant, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "Convert tenant to unstructured")
	}
	unstructuredTenant := &unstructured.Unstructured{Object: objMap}
	newTenant, err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(tenant.Namespace).Update(ctx, unstructuredTenant, metav1.UpdateOptions{})
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

func ListAllOBTenants(ctx context.Context, listOptions metav1.ListOptions) (*v1alpha1.OBTenantList, error) {
	clt := client.GetClient()
	tenantList, err := clt.DynamicClient.Resource(schema.OBTenantRes).List(ctx, listOptions)
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

func GetOBTenant(ctx context.Context, nn types.NamespacedName) (*v1alpha1.OBTenant, error) {
	clt := client.GetClient()
	tenant, err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(nn.Namespace).Get(ctx, nn.Name, metav1.GetOptions{})
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

func DeleteOBTenant(ctx context.Context, nn types.NamespacedName) error {
	clt := client.GetClient()
	err := clt.DynamicClient.Resource(schema.OBTenantRes).Namespace(nn.Namespace).Delete(ctx, nn.Name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "Delete tenant")
	}
	return nil
}

func CreateOBTenantOperation(ctx context.Context, op *v1alpha1.OBTenantOperation) (*v1alpha1.OBTenantOperation, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(op)
	if err != nil {
		logger.Info("Convert tenant operation to unstructured", "err", err)
		return nil, errors.Wrap(err, "Convert tenant operation to unstructured")
	}
	tenantUnstructured := &unstructured.Unstructured{Object: objMap}
	tenantUnstructured.SetGroupVersionKind(schema.OBTenantOperationGVK)
	newTenant, err := clt.DynamicClient.Resource(schema.OBTenantOperationGVR).Namespace(op.Namespace).Create(ctx, tenantUnstructured, metav1.CreateOptions{})
	if err != nil {
		logger.Info("Create tenant operation", "err", err)
		return nil, errors.Wrap(err, "Create tenant ooperation")
	}
	operation := &v1alpha1.OBTenantOperation{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(newTenant.UnstructuredContent(), operation)
	if err != nil {
		logger.Info("Convert unstructured tenant operation to typed", "err", err)
		return nil, errors.Wrap(err, "Convert unstructured tenant operation to typed")
	}
	return operation, nil
}

func GetTenantBackupPolicy(ctx context.Context, nn types.NamespacedName) (*v1alpha1.OBTenantBackupPolicy, error) {
	clt := client.GetClient()
	policy, err := clt.DynamicClient.Resource(schema.OBTenantBackupPolicyGVR).Namespace(nn.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelTenantName + "=" + nn.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get tenant backup policy")
	}
	p := &v1alpha1.OBTenantBackupPolicyList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(policy.UnstructuredContent(), p)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured tenant backup policy to typed")
	}
	if len(p.Items) == 0 {
		return nil, nil
	}
	return &p.Items[0], nil
}

func CreateTenantBackupPolicy(ctx context.Context, policy *v1alpha1.OBTenantBackupPolicy) (*v1alpha1.OBTenantBackupPolicy, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(policy)
	if err != nil {
		logger.Info("Convert tenant backup policy to unstructured", "err", err)
		return nil, errors.Wrap(err, "Convert tenant backup policy to unstructured")
	}
	policyUnstructured := &unstructured.Unstructured{Object: objMap}
	policyUnstructured.SetGroupVersionKind(schema.OBTenantBackupPolicyGVK)
	newPolicy, err := clt.DynamicClient.Resource(schema.OBTenantBackupPolicyGVR).Namespace(policy.Namespace).Create(ctx, policyUnstructured, metav1.CreateOptions{})
	if err != nil {
		logger.Info("Create tenant backup policy", "err", err)
		return nil, errors.Wrap(err, "Create tenant backup policy")
	}
	p := &v1alpha1.OBTenantBackupPolicy{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(newPolicy.UnstructuredContent(), p)
	if err != nil {
		logger.Info("Convert unstructured tenant backup policy to typed", "err", err)
		return nil, errors.Wrap(err, "Convert unstructured tenant backup policy to typed")
	}
	return p, nil
}

func UpdateTenantBackupPolicy(ctx context.Context, policy *v1alpha1.OBTenantBackupPolicy) (*v1alpha1.OBTenantBackupPolicy, error) {
	clt := client.GetClient()
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(policy)
	if err != nil {
		logger.Info("Convert tenant backup policy to unstructured", "err", err)
		return nil, errors.Wrap(err, "Convert tenant backup policy to unstructured")
	}
	policyUnstructured := &unstructured.Unstructured{Object: objMap}
	policyUnstructured.SetGroupVersionKind(schema.OBTenantBackupPolicyGVK)
	newPolicy, err := clt.DynamicClient.Resource(schema.OBTenantBackupPolicyGVR).Namespace(policy.Namespace).Update(ctx, policyUnstructured, metav1.UpdateOptions{})
	if err != nil {
		logger.Info("Update tenant backup policy", "err", err)
		return nil, errors.Wrap(err, "Update tenant backup policy")
	}
	p := &v1alpha1.OBTenantBackupPolicy{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(newPolicy.UnstructuredContent(), p)
	if err != nil {
		logger.Info("Convert unstructured tenant backup policy to typed", "err", err)
		return nil, errors.Wrap(err, "Convert unstructured tenant backup policy to typed")
	}
	return p, nil
}

func DeleteTenantBackupPolicy(ctx context.Context, nn types.NamespacedName) error {
	clt := client.GetClient()
	err := clt.DynamicClient.Resource(schema.OBTenantBackupPolicyGVR).Namespace(nn.Namespace).Delete(ctx, nn.Name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "Delete tenant backup policy")
	}
	return nil
}

func ListBackupJobs(ctx context.Context, listOption metav1.ListOptions) (*v1alpha1.OBTenantBackupList, error) {
	clt := client.GetClient()
	backupJobList, err := clt.DynamicClient.Resource(schema.OBTenantBackupGVR).List(ctx, listOption)
	if err != nil {
		logger.Info("List backup jobs", "err", err)
		return nil, errors.Wrap(err, "List backup jobs")
	}
	list := &v1alpha1.OBTenantBackupList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(backupJobList.UnstructuredContent(), list)
	if err != nil {
		logger.Info("Convert backup jobs", "err", err)
		return nil, errors.Wrap(err, "Convert backup jobs")
	}
	return list, nil
}
