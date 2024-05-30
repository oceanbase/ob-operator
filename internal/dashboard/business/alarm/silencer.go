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

package alarm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/silence"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/errors"
	apimodels "github.com/prometheus/alertmanager/api/v2/models"
	opssilence "github.com/prometheus/alertmanager/api/v2/restapi/operations/silence"
	logger "github.com/sirupsen/logrus"
)

func DeleteSilencer(ctx context.Context, id string) error {
	resp, err := getClient().R().SetContext(ctx).SetHeader("content-type", "application/json").Delete(fmt.Sprintf("%s%s/%s", alarmconstant.AlertManagerAddress, alarmconstant.SingleSilencerUrl, id))
	if err != nil {
		return errors.Wrap(err, errors.ErrExternal, "Delete silencer from alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return errors.Newf(errors.ErrExternal, "Delete silencer got unexpected status: %d", resp.StatusCode())
	}
	return nil
}

func GetSilencer(ctx context.Context, id string) (*silence.SilencerResponse, error) {
	gettableSilencer := apimodels.GettableSilence{}
	resp, err := getClient().R().SetContext(ctx).SetHeader("content-type", "application/json").SetResult(&gettableSilencer).Get(fmt.Sprintf("%s%s/%s", alarmconstant.AlertManagerAddress, alarmconstant.SingleSilencerUrl, id))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Get silencer from alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Get silencer got unexpected status: %d", resp.StatusCode())
	}
	return silence.NewSilencerResponse(&gettableSilencer), nil
}

// TODO: fill in instances and rules to matchers
func CreateOrUpdateSilencer(ctx context.Context, param *silence.SilencerParam) (*silence.SilencerResponse, error) {
	startTime := strfmt.DateTime(time.Now())
	endTime := strfmt.DateTime(time.Unix(param.EndsAt, 0))
	matchers := make(apimodels.Matchers, 0)
	rules := strings.Join(param.Rules, alarmconstant.RegexOR)
	falseValue := false
	trueValue := true
	ruleName := alarmconstant.LabelRuleName
	matchers = append(matchers, &apimodels.Matcher{
		IsEqual: &trueValue,
		IsRegex: &trueValue,
		Name:    &ruleName,
		Value:   &rules,
	})
	instanceType := oceanbase.TypeUnknown
	labelOBCluster := alarmconstant.LabelOBCluster
	labelInstance := alarmconstant.LabelOBCluster
	obcluster := ""
	instances := make([]string, 0, len(param.Instances))
	for _, instance := range param.Instances {
		if instanceType == oceanbase.TypeUnknown {
			instanceType = instance.Type
		}
		if instance.Type != instanceType {
			return nil, errors.New(errors.ErrBadRequest, "All instances should belong to one type")
		}
		if instanceType != oceanbase.TypeOBCluster && obcluster != "" && obcluster != instance.OBCluster {
			return nil, errors.New(errors.ErrBadRequest, "All instances should belong to one obcluster")
		}
		obcluster = instance.OBCluster
		switch instance.Type {
		case oceanbase.TypeOBCluster:
			instances = append(instances, instance.OBCluster)
		case oceanbase.TypeOBServer:
			instances = append(instances, instance.OBServer)
			labelInstance = alarmconstant.LabelOBServer
		case oceanbase.TypeOBZone:
			instances = append(instances, instance.OBZone)
			labelInstance = alarmconstant.LabelOBZone
		case oceanbase.TypeOBTenant:
			instances = append(instances, instance.OBTenant)
			labelInstance = alarmconstant.LabelOBTenant
		default:
			return nil, errors.New(errors.ErrBadRequest, "Unknown instance type")
		}
	}
	instanceValues := strings.Join(instances, alarmconstant.RegexOR)
	if instanceType != oceanbase.TypeOBCluster {
		matchers = append(matchers, &apimodels.Matcher{
			IsEqual: &trueValue,
			IsRegex: &falseValue,
			Name:    &labelOBCluster,
			Value:   &obcluster,
		})
	}
	matchers = append(matchers, &apimodels.Matcher{
		IsEqual: &trueValue,
		IsRegex: &trueValue,
		Name:    &labelInstance,
		Value:   &instanceValues,
	})
	for _, m := range param.Matchers {
		matchers = append(matchers, &apimodels.Matcher{
			IsEqual: &trueValue,
			IsRegex: &m.IsRegex,
			Name:    &m.Name,
			Value:   &m.Value,
		})
	}

	silencer := apimodels.Silence{
		Comment:   &param.Comment,
		CreatedBy: &param.CreatedBy,
		StartsAt:  &startTime,
		EndsAt:    &endTime,
		Matchers:  matchers,
	}
	postableSilence := &apimodels.PostableSilence{
		Silence: silencer,
	}
	okBody := opssilence.PostSilencesOKBody{}
	resp, err := getClient().R().SetContext(ctx).SetHeader("content-type", "application/json").SetBody(postableSilence).SetResult(&okBody).Post(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.MultiSilencerUrl))
	if err != nil || resp.StatusCode() != http.StatusOK {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query silencers from alertmanager")
	}
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Create silencer in alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Create silencer in alertmanager got unexpected status: %d", resp.StatusCode())
	}
	state := string(silence.StateActive)
	gettableSilencer := apimodels.GettableSilence{
		Silence: silencer,
		ID:      &okBody.SilenceID,
		Status: &apimodels.SilenceStatus{
			State: &state,
		},
		UpdatedAt: &startTime,
	}
	silencerResponse := silence.NewSilencerResponse(&gettableSilencer)
	return silencerResponse, nil
}

func ListSilencers(ctx context.Context, filter *silence.SilencerFilter) ([]silence.SilencerResponse, error) {
	gettableSilencers := make(apimodels.GettableSilences, 0)
	req := getClient().R().SetContext(ctx).SetHeader("content-type", "application/json")
	resp, err := req.SetResult(&gettableSilencers).Get(fmt.Sprintf("%s%s", alarmconstant.AlertManagerAddress, alarmconstant.MultiSilencerUrl))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Query silencers from alertmanager")
	} else if resp.StatusCode() != http.StatusOK {
		return nil, errors.Newf(errors.ErrExternal, "Query silencers from alertmanager got unexpected status: %d", resp.StatusCode())
	}
	logger.Infof("resp: %v", resp)
	logger.Infof("silencers: %v", gettableSilencers)
	filteredSilencers := make([]silence.SilencerResponse, 0)
	for _, gettableSilencer := range gettableSilencers {
		silencer := silence.NewSilencerResponse(gettableSilencer)
		if filterSilencer(silencer, filter) {
			filteredSilencers = append(filteredSilencers, *silencer)
		}
	}
	return filteredSilencers, nil
}

func filterSilencer(silencer *silence.SilencerResponse, filter *silence.SilencerFilter) bool {
	matched := true
	if filter.Keyword != "" {
		matched = matched && strings.Contains(silencer.Comment, filter.Keyword)
	}
	// require at least one instance matches
	// TODO: whether to consider a cluster in filter matches a tenant or observer if the cluster names are same
	if filter.Instance != nil {
		instanceMatched := false
		for _, instance := range silencer.Instances {
			if instance.Equals(filter.Instance) {
				instanceMatched = true
				break
			}
		}
		matched = matched && instanceMatched
	}
	return matched
}
