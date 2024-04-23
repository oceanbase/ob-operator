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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

func GetCNIFromAnnotation(pod *corev1.Pod) string {
	_, found := pod.Annotations[oceanbaseconst.AnnotationCalicoValidate]
	if found {
		return oceanbaseconst.CNICalico
	}
	return oceanbaseconst.CNIUnknown
}

func NeedAnnotation(pod *corev1.Pod, cni string) bool {
	switch cni {
	case oceanbaseconst.CNICalico:
		_, found := pod.Annotations[oceanbaseconst.AnnotationCalicoIpAddrs]
		return !found
	default:
		return false
	}
}

// GetTenantRestoreSource gets restore source from tenant CR. If tenantCR is in form of ns/name, the parameter ns is ignored.
func GetTenantRestoreSource(ctx context.Context, clt client.Client, logger *logr.Logger, ns, tenantCR string) (string, error) {
	finalNs := ns
	finalTenantCR := tenantCR
	splits := strings.Split(tenantCR, "/")
	if len(splits) == 2 {
		finalNs = splits[0]
		finalTenantCR = splits[1]
	}
	var restoreSource string
	var err error

	primary := &v1alpha1.OBTenant{}
	err = clt.Get(ctx, types.NamespacedName{
		Namespace: finalNs,
		Name:      finalTenantCR,
	}, primary)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return "", err
		}
	} else {
		obcluster := &v1alpha1.OBCluster{}
		err := clt.Get(ctx, types.NamespacedName{
			Namespace: finalNs,
			Name:      primary.Spec.ClusterName,
		}, obcluster)
		if err != nil {
			return "", errors.Wrap(err, "get obcluster")
		}
		con, err := GetSysOperationClient(clt, logger, obcluster)
		if err != nil {
			return "", errors.Wrap(err, "get oceanbase operation manager")
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
		restoreSource = fmt.Sprintf("SERVICE=%s USER=%s@%s PASSWORD=%s", strings.Join(ipList, ";"), oceanbaseconst.StandbyROUser, primary.Spec.TenantName, standbyRoPwd)
	}

	return restoreSource, nil
}
