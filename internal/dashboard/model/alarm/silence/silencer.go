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

package silence

import (
	"time"

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"

	apimodels "github.com/prometheus/alertmanager/api/v2/models"
	logger "github.com/sirupsen/logrus"
)

type SilencerApiResponse struct {
	Id string `json:"id" binding:"required"`
}

type Status struct {
	State State `json:"state" binding:"required"`
}

type Silencer struct {
	Comment   string          `json:"comment" binding:"required"`
	CreatedBy string          `json:"createdBy" binding:"required"`
	StartsAt  int64           `json:"startsAt" binding:"required"`
	EndsAt    int64           `json:"endsAt" binding:"required"`
	Matchers  []alarm.Matcher `json:"matchers" binding:"required"`
}

type SilencerResponse struct {
	Id        string                 `json:"id" binding:"required"`
	Instances []oceanbase.OBInstance `json:"instance" binding:"required"`
	Status    *Status                `json:"status" binding:"required"`
	UpdatedAt int64                  `json:"updatedAt" binding:"required"`
	Silencer  `json:",inline"`
}

type SilencerIdentity struct {
	Id string `json:"id" binding:"required"`
}

type SilencerParam struct {
	Id        string                 `json:"id,omitempty"`
	Instances []oceanbase.OBInstance `json:"instance" binding:"required"`
	Silencer
}

func extractInstances(matcherMap map[string]alarm.Matcher) []oceanbase.OBInstance {
	instances := make([]oceanbase.OBInstance, 0)
	var matchedInstanceType oceanbase.OBInstanceType
	clusterMatcher, matchCluster := matcherMap[alarmconstant.LabelOBCluster]
	zoneMatcher, matchZone := matcherMap[alarmconstant.LabelOBCluster]
	serverMatcher, matchServer := matcherMap[alarmconstant.LabelOBCluster]
	tenantMatcher, matchTenant := matcherMap[alarmconstant.LabelOBCluster]
	if matchCluster {
		matchedInstanceType = oceanbase.TypeOBCluster
	}
	if matchZone {
		matchedInstanceType = oceanbase.TypeOBZone
	}
	if matchServer {
		matchedInstanceType = oceanbase.TypeOBServer
	}
	if matchTenant {
		matchedInstanceType = oceanbase.TypeOBTenant
	}
	switch matchedInstanceType {
	case oceanbase.TypeOBCluster:
		clusterNames := clusterMatcher.ExtractMatchedValues()
		for _, clusterName := range clusterNames {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBCluster,
				OBCluster: clusterName,
			})
		}
	case oceanbase.TypeOBZone:
		if !matchCluster {
			logger.Error("Cluster matcher not exists")
			break
		} else if clusterMatcher.IsRegex {
			logger.Error("Multiple cluster matches for zone matcher")
			break
		}
		zoneNames := zoneMatcher.ExtractMatchedValues()
		for _, zone := range zoneNames {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBCluster,
				OBCluster: clusterMatcher.Value,
				OBZone:    zone,
			})
		}
	case oceanbase.TypeOBServer:
		if !matchCluster {
			logger.Error("Cluster matcher not exists")
			break
		} else if clusterMatcher.IsRegex {
			logger.Error("Multiple cluster matches for zone matcher")
			break
		}
		serverIps := serverMatcher.ExtractMatchedValues()
		for _, serverIp := range serverIps {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBCluster,
				OBCluster: clusterMatcher.Value,
				OBServer:  serverIp,
			})
		}
	case oceanbase.TypeOBTenant:
		if !matchCluster {
			logger.Error("Cluster matcher not exists")
			break
		} else if clusterMatcher.IsRegex {
			logger.Error("Multiple cluster matches for zone matcher")
			break
		}
		tenantNames := tenantMatcher.ExtractMatchedValues()
		for _, tenant := range tenantNames {
			instances = append(instances, oceanbase.OBInstance{
				Type:      oceanbase.TypeOBCluster,
				OBCluster: clusterMatcher.Value,
				OBTenant:  tenant,
			})
		}
	}
	return instances
}

func NewSilencerResponse(gettableSilencer *apimodels.GettableSilence) *SilencerResponse {
	matchers := make([]alarm.Matcher, 0)
	matcherMap := make(map[string]alarm.Matcher)
	for _, silenceMatcher := range gettableSilencer.Matchers {
		matcher := alarm.Matcher{
			IsRegex: *silenceMatcher.IsRegex,
			Name:    *silenceMatcher.Name,
			Value:   *silenceMatcher.Value,
		}
		matchers = append(matchers, matcher)
		matcherMap[matcher.Name] = matcher
	}

	instances := extractInstances(matcherMap)
	silencer := &Silencer{
		Comment:   *gettableSilencer.Comment,
		CreatedBy: *gettableSilencer.CreatedBy,
		StartsAt:  time.Time(*gettableSilencer.StartsAt).Unix(),
		EndsAt:    time.Time(*gettableSilencer.EndsAt).Unix(),
		Matchers:  matchers,
	}
	silencerResponse := &SilencerResponse{
		Silencer:  *silencer,
		Id:        *gettableSilencer.ID,
		UpdatedAt: time.Time(*gettableSilencer.UpdatedAt).Unix(),
		Status: &Status{
			State: State(*gettableSilencer.Status.State),
		},
		Instances: instances,
	}
	return silencerResponse
}
