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
