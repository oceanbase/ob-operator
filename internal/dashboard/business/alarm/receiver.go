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
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm/receiver"
	"github.com/oceanbase/ob-operator/pkg/errors"
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
