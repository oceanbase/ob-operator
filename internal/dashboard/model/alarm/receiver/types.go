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
package receiver

type ReceiverType string

// TODO add all types
const (
	TypeDingTalk ReceiverType = "dingtalk"
	TypeWeChat                = "wechat"
)

type Receiver struct {
	Name   string       `json:"name" binding:"required"`
	Type   ReceiverType `json:"type" binding:"required"`
	Config string       `json:"config" binding:"required"`
}
