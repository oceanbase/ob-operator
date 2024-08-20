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
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	k8srand "k8s.io/apimachinery/pkg/util/rand"

	apitypes "github.com/oceanbase/ob-operator/api/types"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"

	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

const (
	characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~!@#%^&*_-+=|(){}[]:;,.?/`$\"<>"
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
func GenerateClusterId() int64 {
	return time.Now().Unix() % factor
}
func CheckResourceName(name string) bool {
	// 定义正则表达式
	regex := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`

	// 编译正则表达式
	re, err := regexp.Compile(regex)
	if err != nil {
		log.Println("Error compiling regex:", err)
		return false
	}

	// 检查整个字符串是否符合正则表达式的模式
	return re.MatchString(name)
}
func CheckPassword(password string) bool {
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	for _, char := range password {
		if strings.ContainsRune(characters, char) {
			// 检查字符并增加相应的计数
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
		// 提前返回
		if countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2 {
			return true
		}
	}
	return countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2
}
func MapZonesToTopology(zones map[string]string) ([]param.ZoneTopology, error) {
	if zones == nil {
		return nil, fmt.Errorf("Zone value is required") // 无效的zone信息
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

func GenerateRandomPassword() string {
	const (
		maxLength      = 32
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
		_, err := rand.Read(b) // 随机读取一个字节
		if err != nil {
			panic(err) // 处理错误
		}

		randomIndex := int(b[0]) % len(characters)
		randomChar := characters[randomIndex]
		// 追加字符到密码
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
		if len(sb.String()) >= maxLength {
			return GenerateRandomPassword()
		}
	}

	return sb.String()
}

// GenerateUUID returns uuid
func GenerateUUID() string {
	return k8srand.String(12)
}
