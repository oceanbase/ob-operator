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

package obproxy

const (
	envPrefix = "ODP_"
)

const (
	cmSuffix             = "-odp-config"
	svcSuffix            = "-odp-svc"
	proxyRoSecretSuffix  = "-proxyro-secret"
	proxySysSecretSuffix = "-proxysys-secret"
)

const (
	LabelOBProxy          = "obproxy.oceanbase.com/obproxy"
	LabelWithConfigMap    = "obproxy.oceanbase.com/with-config-map"
	LabelForOBCluster     = "obproxy.oceanbase.com/for-obcluster"
	LabelForNamespace     = "obproxy.oceanbase.com/for-namespace"
	LabelProxyClusterName = "obproxy.oceanbase.com/proxy-cluster-name"
)

const (
	AnnotationServiceType    = "obproxy.oceanbase.com/service-type"
	AnnotationServiceIP      = "obproxy.oceanbase.com/service-ip"
	AnnotationProxySysSecret = "obproxy.oceanbase.com/proxy-sys-secret"
)
