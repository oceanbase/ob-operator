/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	k8srand "k8s.io/apimachinery/pkg/util/rand"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"

	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

const (
	characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~#%^&*_-+|(){}[]:;,.?/\""
	factor     = 4294901759
)

func GenerateUserSecrets(clusterName string, clusterId int64) *apitypes.OBUserSecrets {
	return &apitypes.OBUserSecrets{
		Root:     fmt.Sprintf("%s-%d-root-%s", clusterName, clusterId, GenerateUUID()),
		ProxyRO:  fmt.Sprintf("%s-%d-proxyro-%s", clusterName, clusterId, GenerateUUID()),
		Monitor:  fmt.Sprintf("%s-%d-monitor-%s", clusterName, clusterId, GenerateUUID()),
		Operator: fmt.Sprintf("%s-%d-operator-%s", clusterName, clusterId, GenerateUUID()),
	}
}

// GenerateClusterID generated random cluster ID
func GenerateClusterID() int64 {
	clusterID := time.Now().Unix() % factor
	if clusterID != 0 {
		return clusterID
	}
	return GenerateClusterID()
}

// CheckResourceName checks resource name in k8s
func CheckResourceName(name string) bool {
	regex := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`

	re, err := regexp.Compile(regex)
	if err != nil {
		panic("error when compiling regex expressions")
	}

	return re.MatchString(name)
}

// CheckPassword checks password when creating cluster
func CheckPassword(password string) bool {
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	for _, char := range password {
		if strings.ContainsRune(characters, char) {
			switch {
			case strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ", char):
				countUppercase++
			case strings.ContainsRune("abcdefghijklmnopqrstuvwxyz", char):
				countLowercase++
			case strings.ContainsRune("0123456789", char):
				countNumber++
			default:
				countSpecialChar++
			}
		} else {
			return false
		}
		// if satisfied
		if countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2 {
			return true
		}
	}
	return countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2
}

// CheckTenantName check Tenant name when creating tenant
func CheckTenantName(name string) bool {
	regex := `^[_a-zA-Z][^-]*$`

	re, err := regexp.Compile(regex)
	if err != nil {
		panic("error when compiling regex expressions")
	}

	return re.MatchString(name)
}

// MapZonesToTopology map --zones to zoneTopology
func MapZonesToTopology(zones map[string]string) ([]param.ZoneTopology, error) {
	if zones == nil {
		return nil, fmt.Errorf("Zone replica is required")
	}
	topology := make([]param.ZoneTopology, 0)
	for zoneName, replicaStr := range zones {
		replica, err := strconv.Atoi(replicaStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for zone %s: %s", zoneName, replicaStr)
		}
		topology = append(topology, param.ZoneTopology{
			Zone:         zoneName,
			Replicas:     replica,
			NodeSelector: make([]common.KVPair, 0),
			Tolerations:  make([]common.KVPair, 0),
			Affinities:   make([]common.AffinitySpec, 0),
		})
	}
	return topology, nil
}

// MapZonesToPools map --zones to []resourcePool
func MapZonesToPools(zones map[string]string) ([]param.ResourcePoolSpec, error) {
	if zones == nil {
		return nil, fmt.Errorf("Zone priority is required")
	}
	resourcePool := make([]param.ResourcePoolSpec, 0)
	for zoneName, priorityStr := range zones {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for zone %s: %s", zoneName, priorityStr)
		}
		resourcePool = append(resourcePool, param.ResourcePoolSpec{
			Zone:     zoneName,
			Priority: priority,
			Type:     "Full",
		})
	}
	return resourcePool, nil
}

// MapParameters map --parameters to parameters
func MapParameters(parameters map[string]string) ([]common.KVPair, error) {
	kvMap := make([]common.KVPair, 0)
	for k, v := range parameters {
		kvMap = append(kvMap, common.KVPair{
			Key:   k,
			Value: v,
		})
	}
	return kvMap, nil
}

// GenerateRandomPassword generated random password in range [minLength,maxLength]
func GenerateRandomPassword(minLength int, maxLength int) string {
	const (
		minUppercase   = 2
		minLowercase   = 2
		minNumber      = 2
		minSpecialChar = 2
	)
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	var sb strings.Builder
	for countUppercase < minUppercase || countLowercase < minLowercase || countNumber < minNumber || countSpecialChar < minSpecialChar {
		b := make([]byte, 1)
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		randomIndex := int(b[0]) % len(characters)
		randomChar := characters[randomIndex]
		if err := sb.WriteByte(randomChar); err != nil {
			panic(err)
		}
		switch {
		case strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ", rune(randomChar)):
			countUppercase++
		case strings.ContainsRune("abcdefghijklmnopqrstuvwxyz", rune(randomChar)):
			countLowercase++
		case strings.ContainsRune("0123456789", rune(randomChar)):
			countNumber++
		default:
			countSpecialChar++
		}
	}
	if len(sb.String()) < minLength || len(sb.String()) > maxLength {
		return GenerateRandomPassword(minLength, maxLength)
	}
	return sb.String()
}

// ParseUnitConfig parse param.UnitConfig to v1alpha1.UnitConfig
func ParseUnitConfig(unitConfig *param.UnitConfig) (*v1alpha1.UnitConfig, error) {
	cpuCount, err := resource.ParseQuantity(unitConfig.CPUCount)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid cpu count: " + err.Error())
	}
	memorySize, err := resource.ParseQuantity(unitConfig.MemorySize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid memory size: " + err.Error())
	}
	logDiskSize, err := resource.ParseQuantity(unitConfig.LogDiskSize)
	if err != nil {
		return nil, oberr.NewBadRequest("invalid log disk size: " + err.Error())
	}
	var maxIops, minIops int
	if unitConfig.MaxIops > math.MaxInt32 {
		maxIops = math.MaxInt32
	} else {
		maxIops = int(unitConfig.MaxIops)
	}
	if unitConfig.MinIops > math.MaxInt32 {
		minIops = math.MaxInt32
	} else {
		minIops = int(unitConfig.MinIops)
	}
	return &v1alpha1.UnitConfig{
		MaxCPU:      cpuCount,
		MemorySize:  memorySize,
		MinCPU:      cpuCount,
		LogDiskSize: logDiskSize,
		MaxIops:     maxIops,
		MinIops:     minIops,
		IopsWeight:  unitConfig.IopsWeight,
	}, nil
}

// GenerateUUID returns uuid
func GenerateUUID() string {
	return k8srand.String(12)
}
