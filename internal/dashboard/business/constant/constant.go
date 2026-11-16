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

package constant

import (
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	GB = 1024 * 1024 * 1024
)

const (
	LANGUAGE_ZH_CN = "zh-CN"
	LANGUAGE_EN_US = "en-US"
)

const (
	DASHBOARD_APP_NAME = "oceanbase-dashboard"

	defaultServiceAccountName = "oceanbase-dashboard-sa"
)

// ServiceAccountName returns the service account configured by the deployment.
func ServiceAccountName() string {
	return serviceAccountName(os.Getenv("SERVICE_ACCOUNT"))
}

func serviceAccountName(configuredName string) string {
	if name := strings.TrimSpace(configuredName); name != "" {
		return name
	}
	return defaultServiceAccountName
}

// SharedVolumeSecurityContext keeps report writers on the same SELinux level as
// the Dashboard and cleanup pod. Otherwise each mount can relabel the PVC and
// deny access to the other consumers on enforcing hosts.
func SharedVolumeSecurityContext() *corev1.PodSecurityContext {
	level := strings.TrimSpace(os.Getenv("SHARED_VOLUME_SELINUX_LEVEL"))
	if level == "" {
		return nil
	}
	return &corev1.PodSecurityContext{SELinuxOptions: &corev1.SELinuxOptions{Level: level}}
}
