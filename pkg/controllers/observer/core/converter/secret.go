/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package converter

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerutil "github.com/oceanbase/ob-operator/pkg/controllers/observer/core/util"
)

func GenerateDBUserSecret(obCluster cloudv1.OBCluster, tenantName, userName, password string) corev1.Secret {

	objectMeta := observerutil.GenerateObjectMeta(obCluster, GenerateSecretNameForDBUser(obCluster.Name, tenantName, userName))
	stringData := make(map[string]string)
	stringData["password"] = password
	secret := corev1.Secret{
		ObjectMeta: objectMeta,
		Type:       "Opaque",
		StringData: stringData,
	}
	return secret
}

func GenerateSecretNameForDBUser(clusterName, tenantName, userName string) string {
	return fmt.Sprintf("secret-%s-%s-%s", clusterName, tenantName, userName)
}
