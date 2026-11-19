package obtenantoperation

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	"github.com/oceanbase/ob-operator/pkg/helper"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

type tenantUpgradeClient interface {
	ListTenantWithName(context.Context, string) ([]*model.OBTenant, error)
	SelectCompatibleOfTenants(context.Context) ([]*model.Parameter, error)
	UpgradeTenantWithName(context.Context, string) error
}

func (m *ObTenantOperationManager) upgradePrimaryTenant() error {
	snapshot := m.Resource.Status.PrimaryTenant
	if snapshot == nil || m.Resource.Spec.TargetTenant == nil {
		return errors.New("upgrade target tenant is missing")
	}
	ctx, cancel := context.WithTimeout(m.Ctx, time.Duration(obcfg.GetConfig().Time.DefaultStateWaitTimeout)*time.Second)
	defer cancel()
	const interval = 5 * time.Second
	var tenant v1alpha1.OBTenant
	key := client.ObjectKey{Namespace: m.Resource.Namespace, Name: *m.Resource.Spec.TargetTenant}
	err := wait.PollUntilContextCancel(ctx, interval, true, func(ctx context.Context) (bool, error) {
		if err := m.Client.Get(ctx, key, &tenant); err != nil {
			return false, err
		}
		if (snapshot.UID != "" && tenant.UID != snapshot.UID) || tenant.Spec.TenantName != snapshot.Spec.TenantName || tenant.Spec.ClusterName != snapshot.Spec.ClusterName {
			return false, errors.New("upgrade target tenant was replaced")
		}
		// Nested object metadata is pruned in persisted status snapshots. The
		// operation owner reference retains the original tenant's UID.
		for _, owner := range m.Resource.OwnerReferences {
			if owner.Kind == "OBTenant" && owner.Name == key.Name && owner.UID != tenant.UID {
				return false, errors.New("upgrade target tenant was replaced")
			}
		}
		return tenant.Status.Status == tenantstatus.Running && tenant.Status.TenantRole == constants.TenantRolePrimary, nil
	})
	if err != nil {
		return errors.Wrapf(err, "waiting for primary tenant %s to be running", key)
	}
	con, err := m.getClusterSysClient(tenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	recordID := int64(tenant.Status.TenantRecordInfo.TenantID)
	oldID := int64(snapshot.Status.TenantRecordInfo.TenantID)
	if recordID == 0 {
		recordID = oldID
	} else if oldID > 0 && oldID != recordID {
		return errors.New("upgrade target database tenant ID has changed")
	}
	return upgradeAndWait(ctx, con, tenant.Spec.TenantName, recordID, interval)
}

// The database assigns a tenant ID during restore. A previously accepted
// operation can hold a snapshot from before that ID was available.
func upgradeAndWait(ctx context.Context, db tenantUpgradeClient, name string, tenantID int64, interval time.Duration) error {
	submitted := false
	var submitErr error
	err := wait.PollUntilContextCancel(ctx, interval, true, func(ctx context.Context) (bool, error) {
		rows, err := db.ListTenantWithName(ctx, name)
		if err != nil {
			return false, err
		}
		if len(rows) != 1 || rows[0].TenantID <= 1 || rows[0].Status != "NORMAL" || rows[0].TenantRole != string(constants.TenantRolePrimary) {
			return false, nil
		}
		if tenantID != 0 && tenantID != rows[0].TenantID {
			return false, fmt.Errorf("physical ID for tenant %s changed from %d to %d", name, tenantID, rows[0].TenantID)
		}
		tenantID = rows[0].TenantID
		params, err := db.SelectCompatibleOfTenants(ctx)
		if err != nil {
			return false, err
		}
		atTarget, canSubmit, err := checkUpgradeCompatibility(params, tenantID)
		if err != nil || atTarget {
			return atTarget, err
		}
		if canSubmit && !submitted {
			submitted = true
			// A timeout or a concurrent request can leave a real job running.
			// Observe its outcome before reporting an ambiguous submission as failed.
			submitErr = db.UpgradeTenantWithName(ctx, name)
		}
		return false, nil
	})
	if err != nil && submitErr != nil {
		return errors.Wrapf(submitErr, "upgrade for tenant %s did not converge: %v", name, err)
	}
	return errors.Wrapf(err, "waiting for tenant %s (ID %d) compatibility to converge", name, tenantID)
}

// GV$OB_PARAMETERS can return one row per Observer. Issue the upgrade only
// when all reported target rows are older, and finish only when all match sys.
func checkUpgradeCompatibility(rows []*model.Parameter, tenantID int64) (atTarget, canSubmit bool, err error) {
	var sys *helper.OceanBaseVersion
	var targets []*helper.OceanBaseVersion
	for _, row := range rows {
		if row.TenantID != 1 && row.TenantID != tenantID {
			continue
		}
		v, err := helper.ParseOceanBaseVersion(row.Value)
		if err != nil {
			return false, false, errors.Wrapf(err, "invalid compatible for tenant %d", row.TenantID)
		}
		if row.TenantID == tenantID {
			targets = append(targets, v)
		} else if sys != nil && sys.Cmp(v) != 0 {
			return false, false, nil
		} else {
			sys = v
		}
	}
	if sys == nil || len(targets) == 0 {
		return false, false, nil
	}
	if sys.Cmp(&helper.OceanBaseVersion{Major: 4, Minor: 1}) < 0 {
		return false, false, errors.New("tenant upgrade requires cluster version 4.1.0.0 or later")
	}
	matched := 0
	for _, v := range targets {
		switch cmp := v.Cmp(sys); {
		case cmp > 0:
			return false, false, fmt.Errorf("target tenant compatibility %s exceeds cluster compatibility %s", v, sys)
		case cmp == 0:
			matched++
		}
	}
	return matched == len(targets), matched == 0, nil
}
