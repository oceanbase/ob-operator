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
	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/generated/bindata"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/receiver"
	"github.com/oceanbase/ob-operator/pkg/errors"

	"gopkg.in/yaml.v2"
)

func GetReceiver(name string) (*receiver.Receiver, error) {
	receivers, err := ListReceivers()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Failed to get receivers")
	}
	for _, receiver := range receivers {
		if receiver.Name == name {
			return &receiver, nil
		}
	}
	return nil, errors.NewNotFound("Receiver not found")
}

func ListReceivers() ([]receiver.Receiver, error) {
	config, err := GetAlertmanagerConfig()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrExternal, "Failed to get config")
	}
	receivers := make([]receiver.Receiver, 0, len(config.Receivers))
	for _, amreceiver := range config.Receivers {
		receiver := receiver.NewReceiver(&amreceiver)
		if receiver != nil {
			receivers = append(receivers, *receiver)
		}
	}
	return receivers, nil
}

func ListReceiverTemplates() ([]receiver.Template, error) {
	receiverTemplates := make([]receiver.Template, 0)
	configFile := alarmconstant.ReceiverConfigTemplateFile
	receiverTemplatesConfigContent, err := bindata.Asset(configFile)
	if err != nil {
		return receiverTemplates, errors.Wrap(err, errors.ErrInternal, "Read receiver templates failed")
	}
	err = yaml.Unmarshal(receiverTemplatesConfigContent, &receiverTemplates)
	if err != nil {
		return receiverTemplates, errors.Wrap(err, errors.ErrInternal, "Decode receiver templates failed")
	}
	return receiverTemplates, err
}

func GetReceiverTemplates(receiverType string) (*receiver.Template, error) {
	receiverTemplates, err := ListReceiverTemplates()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternal, "Get receiver templates failed")
	}
	for _, receiverTemplate := range receiverTemplates {
		if receiverTemplate.Type == receiver.ReceiverType(receiverType) {
			return &receiverTemplate, nil
		}
	}
	return nil, errors.NewNotFound("Template for receiver not found")
}
