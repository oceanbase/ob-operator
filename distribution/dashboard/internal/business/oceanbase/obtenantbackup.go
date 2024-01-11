package oceanbase

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"k8s.io/apimachinery/pkg/types"
)

func CreateTenantBackupPolicy(nn types.NamespacedName, p *param.CreateBackupPolicy) (*v1alpha1.OBTenantBackupPolicy, error) {
	return nil, nil
}

func UpdateTenantBackupPolicy(nn types.NamespacedName, p *param.UpdateBackupPolicy) (*v1alpha1.OBTenantBackupPolicy, error) {
	return nil, nil
}
