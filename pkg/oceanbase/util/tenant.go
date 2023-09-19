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

package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
)

const (
	ReplicaPattern = "([a-zA-Z]+)\\{([\\d]+)\\}@([\\w]+)"
)

func ConvertFromReplicaStr(replica string) *model.Replica {
	p := regexp.MustCompile(ReplicaPattern)
	replicaParts := p.FindStringSubmatch(replica)
	num, _ := strconv.Atoi(replicaParts[2])
	if len(replicaParts) == 4 {
		return &model.Replica{
			Type: replicaParts[1],
			Num:  num,
			Zone: replicaParts[3],
		}
	}
	return nil
}

func ConvertToReplicaStr(replica *model.Replica) string {
	return fmt.Sprintf("%s{%d}@%s", replica.Type, replica.Num, replica.Zone)
}

func ConvertFromLocalityStr(locality string) []model.Replica {
	replicas := make([]model.Replica, 0)
	parts := strings.Split(locality, ",")
	for _, p := range parts {
		replica := ConvertFromReplicaStr(strings.TrimSpace(p))
		if replica != nil {
			replicas = append(replicas, *replica)
		}
	}
	return replicas
}

func ConvertToLocalityStr(replicas []model.Replica) string {
	replicaStrs := make([]string, 0)
	for _, replica := range replicas {
		replicaStrs = append(replicaStrs, ConvertToReplicaStr(&replica))
	}
	return strings.Join(replicaStrs, ", ")
}

func OmitZoneFromReplicas(replicas []model.Replica, zone string) []model.Replica {
	result := make([]model.Replica, 0)
	for _, replica := range replicas {
		if replica.Zone != zone {
			result = append(result, replica)
		}
	}
	return result
}
