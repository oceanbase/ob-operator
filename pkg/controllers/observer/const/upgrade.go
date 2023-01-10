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

package observerconst

const (
	UpgradingPassed = "Upgrading Passed"
)

const (
	UpgradePreChecker      = "pre-checker"
	UpgradePostChecker     = "post-checker"
	UpgradePre             = "pre"
	UpgradePost            = "post"
	UpgradePreCheckerPath  = "/home/admin/oceanbase/etc/upgrade_checker.py"
	UpgradeScriptsPath     = "/home/admin/oceanbase/scripts/"
	PreScriptFile          = "/upgrade_pre.py"
	PostScriptFile         = "/upgrade_post.py"
	UpgradePostCheckerPath = "/home/admin/oceanbase/etc/upgrade_post_checker.py"
)

const (
	MinObserverVersion  = "min_observer_version"
	EnableUpgradeMode   = "enable_upgrade_mode"
	ConfigAdditionalDir = "config_additional_dir"
)

const (
	JobRunning   = "Running"
	JobSucceeded = "Succeeded"
	JobFailed    = "Failed"
)
