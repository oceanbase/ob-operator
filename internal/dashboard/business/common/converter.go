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

package common

import (
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
)

func KVsToMap(kvs []modelcommon.KVPair) map[string]string {
	ret := make(map[string]string)
	for _, kv := range kvs {
		ret[kv.Key] = kv.Value
	}
	return ret
}

func MapToKVs(m map[string]string) []modelcommon.KVPair {
	kvs := make([]modelcommon.KVPair, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, modelcommon.KVPair{
			Key:   k,
			Value: v,
		})
	}
	return kvs
}
