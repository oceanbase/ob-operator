/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"k8s.io/apimachinery/pkg/types"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
)

var (
	operatorCheckCmd             = "kubectl get crds -A -o name | grep oceanbase.oceanbase.com"
	certManagerCheckCmd          = "kubectl get crds -o name | grep cert-manager"
	dashboardCheckCmd            = "helm list | grep oceanbase-dashboard"
	localPathProvisionerCheckCmd = "kubectl get deployment -A | grep local-path-provisioner"
)

var (
	// Define the resources to check for each command
	certManagerResources = []string{
		"challenges.acme.cert-manager.io",
		"orders.acme.cert-manager.io",
		"certificaterequests.cert-manager.io",
		"certificates.cert-manager.io",
		"clusterissuers.cert-manager.io",
		"issuers.cert-manager.io",
	}

	operatorResources = []string{
		"obparameters.oceanbase.oceanbase.com",
		"observers.oceanbase.oceanbase.com",
		"obclusters.oceanbase.oceanbase.com",
		"obtenantbackups.oceanbase.oceanbase.com",
		"obtenantrestores.oceanbase.oceanbase.com",
		"obzones.oceanbase.oceanbase.com",
		"obtenants.oceanbase.oceanbase.com",
		"obtenantoperations.oceanbase.oceanbase.com",
		"obtenantbackuppolicies.oceanbase.oceanbase.com",
	}

	dashboardResources = "oceanbase-dashboard"

	localPathProvisionerResources = "local-path-provisioner"
)

// CheckIfClusterExists checks if cluster exists in the environment
func CheckIfClusterExists(ctx context.Context, name string, namespace string) bool {
	cluster, _ := clients.GetOBCluster(ctx, namespace, name)
	return cluster != nil
}

// CheckIfTenantExists checks if tenant exists in the environment
func CheckIfTenantExists(ctx context.Context, name string, namespace string) bool {
	nn := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	tenant, _ := clients.GetOBTenant(ctx, nn)
	return tenant != nil
}

// CheckTenantStatus checks running status of obtenant
func CheckTenantStatus(tenant *v1alpha1.OBTenant) error {
	if tenant.Status.Status != tenantstatus.Running {
		return fmt.Errorf("OBTenant status invalid, Status:%s", tenant.Status.Status)
	}
	return nil
}

// CheckClusterStatus checks running status of obcluster
func CheckClusterStatus(cluster *v1alpha1.OBCluster) error {
	if cluster.Status.Status != clusterstatus.Running {
		return fmt.Errorf("OBCluster status invalid, Status:%s", cluster.Status.Status)
	}
	return nil
}

// CheckPrimaryTenant checks primary tenant for a standbytenant
func CheckPrimaryTenant(standbytenant *v1alpha1.OBTenant) error {
	if standbytenant.Spec.Source == nil || standbytenant.Spec.Source.Tenant == nil {
		return fmt.Errorf("OBTenant %s has no primary tenant", standbytenant.Name)
	}
	return nil
}

// CheckTenantRole checks tenant role
func CheckTenantRole(tenant *v1alpha1.OBTenant, role apitypes.TenantRole) error {
	if tenant.Status.TenantRole != role {
		return fmt.Errorf("Tenant is not %s tenant", string(role))
	}
	return nil
}

// CheckIfComponentExists checks if component exists in the environment
func CheckIfComponentExists(component string) bool {
	switch component {
	case "cert-manager":
		return checkIfResourceExists(certManagerCheckCmd, certManagerResources...)
	case "ob-operator":
		return checkIfResourceExists(operatorCheckCmd, operatorResources...)
	case "ob-dashboard":
		return checkIfResourceExists(dashboardCheckCmd, dashboardResources)
	case "local-path-provisioner":
		return checkIfResourceExists(localPathProvisionerCheckCmd, localPathProvisionerResources)
	default:
		return false
	}
}

// checkIfResourceExists checks if the resource exists in the environment
func checkIfResourceExists(checkCmd string, resourceList ...string) bool {
	cmd := exec.Command("sh", "-c", checkCmd)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false
	}

	output := out.Bytes()
	for _, resource := range resourceList {
		if !bytes.Contains(output, []byte(resource)) {
			return false
		}
	}
	return true
}
