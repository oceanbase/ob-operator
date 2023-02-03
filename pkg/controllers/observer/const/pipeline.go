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

import (
	"time"
)

// scale state machine
const (
	ScaleUP   = "Scale UP"
	ScaleDown = "Scale Down"
	Maintain  = "Maintain"
)

// zone scale state machine
const (
	ZoneScaleUP   = "Zone Scale UP"
	ZoneScaleDown = "Zone Scale Down"
	ZoneMaintain  = "Zone Maintain"
)

// obcluster upgrade state machine
const (
	UpgradeModeBP    = "BP"
	NeedUpgradeCheck = "Need Upgrade Check"
	UpgradeChecking  = "Upgrade Checking"

	NeedExecutingPreScripts = "Need Executing Pre Scripts"
	ExecutingPreScripts     = "Executing Pre Scripts"

	NeedUpgrading = "Need Upgrading"
	Upgrading     = "Upgrading"

	ExecutingPostScripts = "Executing Post Scripts"

	NeedUpgradePostCheck = "Need Upgrade Post Check"
	UpgradePostChecking  = "Upgrade Post Checking"
)

// Step
const (
	StepBootstrap = "Bootstrap"
	StepMaintain  = "Maintain"
)

const (
	BootstrapTimeout = 600

	AddServerTimeout                 = 60
	TickPeriodForOBServerStatusCheck = 5 * time.Second
	TickNumForOBServerStatusCheck    = 60

	TickPeriodForRSJobStatusCheck = 5 * time.Second
	TickNumForRSJobStatusCheck    = 12

	TickPeriodForPodStatusCheck = 1 * time.Second
	TickNumForPodStatusCheck    = 60

	DelServerTimeout = 30
)
