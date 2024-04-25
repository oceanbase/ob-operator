/*
Copyright (c) 2023 OceanBase
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
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

// GetTenantRestoreSource gets restore source from tenant CR. If tenantCR is in form of ns/name, the parameter ns is ignored.
func GetTenantRestoreSource(ctx context.Context, clt client.Client, logger *logr.Logger, ns, tenantCR string) (string, error) {
	primary, con, err := getTenantAndClusterCon(ctx, clt, logger, ns, tenantCR)
	if err != nil {
		return "", err
	}
	// Get ip_list from primary tenant
	aps, err := con.ListTenantAccessPoints(primary.Spec.TenantName)
	if err != nil {
		return "", err
	}
	ipList := make([]string, 0)
	for _, ap := range aps {
		ipList = append(ipList, fmt.Sprintf("%s:%d", ap.SvrIP, ap.SqlPort))
	}
	standbyRoPwd, err := ReadPassword(clt, ns, primary.Status.Credentials.StandbyRO)
	if err != nil {
		logger.Error(err, "Failed to read standby ro password")
		return "", err
	}
	// Set restore source
	restoreSource := fmt.Sprintf("SERVICE=%s USER=%s@%s PASSWORD=%s", strings.Join(ipList, ";"), oceanbaseconst.StandbyROUser, primary.Spec.TenantName, standbyRoPwd)

	return restoreSource, nil
}

// CheckTenantLSIntegrity checks LS integrity of tenant CR. If tenantCR is in form of ns/name, the parameter ns is ignored.
func CheckTenantLSIntegrity(ctx context.Context, clt client.Client, logger *logr.Logger, ns, tenantCR string) error {
	primary, con, err := getTenantAndClusterCon(ctx, clt, logger, ns, tenantCR)
	if err != nil {
		return err
	}
	// Check LS integrity
	lsDeletion, err := con.ListLSDeletion(int64(primary.Status.TenantRecordInfo.TenantID))
	if err != nil {
		return err
	}
	if len(lsDeletion) > 0 {
		return errors.New("LS deletion set is not empty, log is of not integrity")
	}
	logStats, err := con.ListLogStats(int64(primary.Status.TenantRecordInfo.TenantID))
	if err != nil {
		return err
	}
	if len(logStats) == 0 {
		return errors.New("Log stats is empty, out of expectation")
	}
	for _, ls := range logStats {
		if ls.BeginLSN != 0 {
			return errors.New("Log stats begin SCN is not 0, log is of not integrity")
		}
	}

	return nil
}

func getTenantAndClusterCon(ctx context.Context, clt client.Client, logger *logr.Logger, ns, tenantCR string) (*v1alpha1.OBTenant, *operation.OceanbaseOperationManager, error) {
	finalNs := ns
	finalTenantCR := tenantCR
	splits := strings.Split(tenantCR, "/")
	if len(splits) == 2 {
		finalNs = splits[0]
		finalTenantCR = splits[1]
	}
	var err error
	primary := &v1alpha1.OBTenant{}
	err = clt.Get(ctx, types.NamespacedName{
		Namespace: finalNs,
		Name:      finalTenantCR,
	}, primary)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, nil, err
		}
		return nil, nil, errors.New("tenant not found")
	} else {
		obcluster := &v1alpha1.OBCluster{}
		err := clt.Get(ctx, types.NamespacedName{
			Namespace: finalNs,
			Name:      primary.Spec.ClusterName,
		}, obcluster)
		if err != nil {
			return primary, nil, errors.Wrap(err, "get obcluster")
		}
		con, err := GetSysOperationClient(clt, logger, obcluster)
		if err != nil {
			return primary, nil, errors.Wrap(err, "get oceanbase operation manager")
		}
		return primary, con, nil
	}
}
