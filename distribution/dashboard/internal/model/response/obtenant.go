package response

import "github.com/oceanbase/ob-operator/api/v1alpha1"

type OBTenant struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	CreatedAt string `json:"createdAt"`

	Spec   v1alpha1.OBTenantSpec   `json:"spec"`
	Status v1alpha1.OBTenantStatus `json:"status"`
}
