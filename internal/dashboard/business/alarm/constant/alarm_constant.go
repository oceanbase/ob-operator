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

const (
	DefaultAlarmQueryTimeout = 5
)

const (
	ReceiverConfigTemplateFile = "internal/assets/receiver_templates.yaml"
)

const (
	AlertManagerAddress   = "http://127.0.0.1:9093"
	AlertUrl              = "/api/v2/alerts"
	SingleSilencerUrl     = "/api/v2/silence"
	MultiSilencerUrl      = "/api/v2/silences"
	RuleUrl               = "/api/v1/rules"
	PrometheusReloadUrl   = "/-/reload"
	AlertmanagerReloadUrl = "/-/reload"
	StatusUrl             = "/api/v2/status"
)

const (
	LabelOBCluster = "ob_cluster_name"
	LabelOBZone    = "obzone"
	LabelOBServer  = "svr_ip"
	LabelOBTenant  = "tenant_name"
)

const (
	LabelRuleName     = "rule_name"
	LabelServerity    = "serverity"
	LabelInstanceType = "instance_type"
)

const (
	AnnoSummary     = "summary"
	AnnoDescription = "description"
)

const (
	OBRuleGroupName = "ob-rule"
)

const (
	RuleConfigFile = "/etc/prometheus/rules/prometheus.rules"
)

const (
	AlertmanagerConfigFile = "/etc/alertmanager/alertmanager.yaml"
)

const (
	RegexOR = "|"
)
